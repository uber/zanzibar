#!/usr/bin/env node
'use strict';

var fs = require('fs');
var path = require('path');
var assert = require('assert');

function GocovParser(gocovObj) {
    this.input = gocovObj;
    this.fileLocCache = Object.create(null);
    this.output = Object.create(null);
}

GocovParser.prototype.parse = function parse() {
    for (var i = 0; i < this.input.Packages.length; i++) {
        var p = this.input.Packages[i];
        var name = p.Name;
        var functions = p.Functions;

        for (var j = 0; j < functions.length; j++) {
            this.parseFunction(name, functions[j]);
        }
    }
};

GocovParser.prototype.parseFunction =
function parseFunction(folderName, functionInfo) {
    // console.error('functionInfo', functionInfo);
    var filePath = path.join(process.cwd(), functionInfo.File);

    var fileObj = this.output[filePath];
    if (!fileObj) {
        fileObj = this.output[filePath] = new FileCoverageInfo(filePath);
    }

    var fileLoc = this.fileLocCache[filePath];
    if (!fileLoc) {
        fileLoc = this.fileLocCache[filePath] = new FileLoc(filePath);
    }

    var fnName = functionInfo.Name;
    var startOffset = functionInfo.Start;
    var endOffset = functionInfo.End;

    var startLoc = fileLoc.computeFunctionStartLocation(startOffset);
    var endLoc = fileLoc.computeFunctionEndLocation(endOffset);

    // console.error('functionLoc', {
    //     start: startLoc,
    //     end: endLoc
    // });

    var fnId = fileObj.fnCounter;
    fileObj.f[fnId] = 1
    fileObj.fnMap[fnId] = {
        name: fnName,
        line: startLoc.line,
        loc: {
            start: startLoc,
            end: endLoc
        },
        skip: false
    };
    fileObj.fnCounter++;

    // TODO(sindelar): Investigate
    if (functionInfo.Statements == null) {
        return
    }
    for (var i = 0; i < functionInfo.Statements.length; i++) {
        var statement = functionInfo.Statements[i];
        
        var startLoc = fileLoc.computeStatementLocation(statement.Start);
        var endLoc = fileLoc.computeStatementLocation(statement.End);

        var skipped = false
            
        // ignoredLines is 0 indexed, startLoc is 1 indexed
        var lineIndex = startLoc.line - 1;
        if (fileLoc.ignoredLines.indexOf(lineIndex) >= 0) {
            // If this statement is ignored then it is ignored
            skipped = true
        }

        var sId = fileObj.sCounter;
        fileObj.s[sId] = statement.Reached;
        fileObj.statementMap[sId] = {
            start: startLoc,
            end: endLoc,
            skip: skipped
        };

        // console.error('handleStatement', {
        //     fnName: fnName,
        //     offsets: [statement.Start,statement.End],
        //     fileName: filePath,
        //     start: startLoc,
        //     end: endLoc,
        //     ignoredLines: fileLoc.ignoredLines
        // });

        fileObj.sCounter++;
    }
};

function FileCoverageInfo(filePath) {
    this.path = filePath;
    // statement counts, st id -> count
    this.s = Object.create(null);
    this.sCounter = 1;
    // branch counts, br id -> count
    this.b = Object.create(null);
    // function counts, fn id -> count
    this.f = Object.create(null);
    this.fnCounter = 1;
    // fnMap, fn id -> FnInfo{name, line, loc, skip}
    this.fnMap = Object.create(null);
    // statementMap, st id -> Loc{start, end}
    this.statementMap = Object.create(null);
    // branchMap, st id -> BrInfo{line, type, locations}
    this.branchMap = Object.create(null);
}

function FileLoc(fileName) {
    this.fileName = fileName;
    this.fileContent = fs.readFileSync(fileName, 'utf8');

    this.lines = this.fileContent.split('\n');

    this.lineStarts = [];
    var soFar = 0;
    for (var i = 0; i < this.lines.length; i++) {
        this.lineStarts[i] = soFar;
        soFar += this.lines[i].length + 1;
    }

    // 0 indexed...
    this.ignoredLines = [];
    for (var i = 0; i < this.lines.length; i++) {
        if (this.lines[i].indexOf('coverage ignore next line') >= 0) {
            this.ignoredLines.push(i + 1);
        }
    }
}

FileLoc.prototype.computeStatementLocation =
function computeStatementLocation(offset) {
    var text = this.fileContent.slice(0, offset);
    var lineNo = text.split('\n').length;

    var lineStart = this.lineStarts[lineNo - 1];
    // console.error('info', lineNo, lineStart, offset - lineStart);

    return {
        line: lineNo,
        column: offset - lineStart
    };
}

FileLoc.prototype.computeFunctionStartLocation =
function computeFunctionStartLocation(offset) {
    var text = this.fileContent.slice(0, offset);
    var lineNo = text.split('\n').length;
    assert(lineNo > 0, "lineNo must not be zero...");

    return {
        line: lineNo,
        column: 0
    };
};

FileLoc.prototype.computeFunctionEndLocation =
function computeFunctionEndLocation(offset) {
    var text = this.fileContent.slice(0, offset);
    var lineNo = text.split('\n').length;
    assert(lineNo > 0, "lineNo must not be zero...");

    return {
        line: lineNo,
        column: this.lines[lineNo - 1].length
    };
};

function main() {
    var fileName = process.argv[2] || '';

    var isExists = fs.existsSync(fileName);
    if (!isExists) {
        console.error('./gocov-to-istanbul-coverage.js [gocov.json]')
        console.error('Cannot find file $1')
        console.error('')
        process.exit(1)
        return;
    }

    var text = fs.readFileSync(fileName, 'utf8');
    var gocov = JSON.parse(text);

    var parser = new GocovParser(gocov);
    parser.parse();

    console.log(JSON.stringify(parser.output));
}

if (require.main === module) {
    main();
}
