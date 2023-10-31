# Fastly's Next-Gen WAF Terraform Configuration generation tool

What does this thing do?

It imports your NGWAF settings and generates the Terraform configuration for your NGWAF config in HCL.

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