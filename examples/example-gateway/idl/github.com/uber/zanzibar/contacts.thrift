namespace java com.uber.zanzibar.services.contacts

include "./exception.thrift"

typedef string UUID
typedef string ContactFragmentType

struct ContactFragment {
    1: optional ContactFragmentType type
    2: optional string text
}

struct ContactAttributes {
    1: optional string firstName
    2: optional string lastName
    3: optional string nickname
    4: optional bool hasPhoto
    5: optional i32 numFields
    6: optional i32 timesContacted
    7: optional i32 lastTimeContacted
    8: optional bool isStarred
    9: optional bool hasCustomRingtone
    10: optional bool isSendToVoicemail
    11: optional bool hasThumbnail
    12: optional string namePrefix
    13: optional string nameSuffix
}

struct Contact {
    1: optional list<ContactFragment> fragments
    2: optional ContactAttributes attributes
}

struct SaveContactsRequest {
    1: required UUID userUUID (zanzibar.http.ref = "params.userUUID");
    2: required list<Contact> contacts
}

struct SaveContactsResponse {
}

service Contacts {
    SaveContactsResponse saveContacts(
        1: required SaveContactsRequest saveContactsRequest
    ) throws (
        1: exception.TemporaryRedirect temporaryRedirect (zanzibar.http.status = "307")
        2: exception.Unauthenticated unauthenticated (zanzibar.http.status = "401")
        3: exception.BadRequest badRequest (zanzibar.http.status = "422")
        4: exception.ServerError serverError (zanzibar.http.status = "500")
    ) (
        zanzibar.http.method = "POST"
        zanzibar.http.path = "/contacts/:userUUID/contacts"
        zanzibar.http.status = "202"
        zanzibar.http.req.def.boxed = "true"
        zanzibar.meta = "TokenAuth,Client"
        zanzibar.handler = "contacts.saveContacts"
    )
}