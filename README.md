# Fastly's Next-Gen WAF Terraform Configuration generation tool

It imports your NGWAF settings and generates the Terraform configuration for your NGWAF config in HCL.

## Feature list and status
- [x] Corp Rules                :white_check_mark:
- [x] Corp Lists                :white_check_mark:
- [x] Corp Signals              :white_check_mark:
- [x] Site Request Rules        :white_check_mark:
- [x] Site Rate Limiting Rules  :white_check_mark:
- [x] Site Templated Rules      :white_check_mark:
- [x] Site Lists                :white_check_mark:
- [x] Site Signals              :white_check_mark:
- [x] Site Alerts               :white_check_mark:
- [x] Site Agent Alerts         :white_check_mark:
- [ ] Space lasers
- [ ] Coffee maker

# Set up
Environment variables must exist
```
TF_VAR_NGWAF_CORP
TF_VAR_NGWAF_SITE
TF_VAR_NGWAF_EMAIL
TF_VAR_NGWAF_TOKEN
```

Just run `make run`


# Need to start over?
`make rerun`

# Limitations
There are many limitations. This currently only attempts to import site request rules.
