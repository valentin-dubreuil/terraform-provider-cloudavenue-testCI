---
page_title: "cloudavenue_publicip Resource - cloudavenue"
subcategory: "Public IP"
description: |-
  The public IP resource allows you to manage a public IP on your Organization.
---

# cloudavenue_publicip (Resource)

The public IP resource allows you to manage a public IP on your Organization.

## Example Usage

```terraform
data "cloudavenue_edgegateways" "example" {}

resource "cloudavenue_publicip" "example" {
  edge_id = data.cloudavenue_edgegateways.example.edge_gateways[0].edge_id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `edge_id` (String) The ID of the Edge Gateway.
- `edge_name` (String) The name of the Edge Gateway.
- `timeouts` (Attributes) (see [below for nested schema](#nestedatt--timeouts))
- `vdc` (String) Public IP is natted toward the INET VDC Edge in the specified VDC Name. This parameter helps to find target VDC Edge in case of multiples INET VDC Edges with same names

### Read-Only

- `id` (String) The ID is the public IP address.
- `public_ip` (String) Public IP address.

<a id="nestedatt--timeouts"></a>
### Nested Schema for `timeouts`

Optional:

- `create` (String)
- `delete` (String)
- `read` (String)

## Import

Import is supported using the following syntax:
```shell
# use the `id` to import an existing public IP
# `id` is the public IP
terraform import cloudavenue_publicip.example <id>
```