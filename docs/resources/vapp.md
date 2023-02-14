---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "cloudavenue_vapp Resource - cloudavenue"
subcategory: ""
description: |-
  The Edge Gateway resource allows you to create and manage Edge Gateways in CloudAvenue.
---

# cloudavenue_vapp (Resource)

The Edge Gateway resource allows you to create and manage Edge Gateways in CloudAvenue.

## Example Usage

```terraform
resource "cloudavenue_vapp" "example" {
  vapp_name = "vapp_name"
  description = "This is a test vapp"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `vapp_name` (String) A name for the vApp, unique within the VDC. Required if `vapp_id` is not set.

### Optional

- `description` (String) Optional description of the vApp
- `guest_properties` (Map of String) Key/value settings for guest properties
- `lease` (Block List) Defines lease parameters for this vApp (see [below for nested schema](#nestedblock--lease))
- `power_on` (Boolean) A boolean value stating if this vApp should be powered on
- `vdc` (String) The name of VDC to use, optional if defined at provider level

### Read-Only

- `href` (String) vApp Hyper Reference
- `id` (String) The ID is a `vapp_id`.
- `status_code` (Number) Shows the status code of the vApp
- `status_text` (String) Shows the status of the vApp
- `vapp_id` (String) The ID of vApp

<a id="nestedblock--lease"></a>
### Nested Schema for `lease`

Optional:

- `runtime_lease_in_sec` (Number) How long any of the VMs in the vApp can run before the vApp is automatically powered off or suspended. 0 means never expires. Max value is 3600
- `storage_lease_in_sec` (Number) How long the vApp is available before being automatically deleted or marked as expired. 0 means never expires. Max value is 3600

## Import

Import is supported using the following syntax:

```shell
# use the public ip to import the public ip
terraform import cloudavenue_vapp.example vapp_name
```