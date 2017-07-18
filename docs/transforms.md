
Examples of mappings and overriding of fields based on the types of the
ToField, FromField, OverriddenField and whether the Override flag is set in the transform.
The ToField is the output of the transform, the OverriddenField is the default input field,
and the FromField is the input field defined in the transform.

For legacy reasons, we support the Override Flag in transforms to allow endpoint owners
to configure whether the FromFeild should overwrite the OverriddenField when it is present.

| ToFieldOpt | FromFieldOpt | OverriddenFieldOpt | OverrideFlag | Code |
| :----------:|:--------------:|:----------:|:---------:|:---------|
| Optional      | Optional | Optional | false | := overridden \| from
| Optional      | Optional | Optional | true | := from \| overridden
| Optional      | Optional | Required | false | := overridden
| Optional      | Optional | Required | true | := from \| overridden
| Optional      | Required | Optional | false | :=  overridden \|from
| Optional      | Required | Optional | true | := overridden \| from
| Optional      | Required | Required | false | := from
| Optional      | Required | Required | true | := overridden
| Required      | Optional | Optional | false | := overridden \| from \| new
| Required      | Optional | Optional | true | := from \| overridden \| new
| Required      | Optional | Required | false | := overridden
| Required      | Optional | Required | true | := from \| overridden
| Required      | Required | Optional | false | :=  overridden \|from
| Required      | Required | Optional | true | := from
| Required      | Required | Required | false | := overridden
| Required      | Required | Required | true | := from