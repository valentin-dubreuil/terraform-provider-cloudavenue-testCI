---
page_title: "cloudavenue_vm_disk Resource - cloudavenue"
subcategory: "VM (Virtual Machine)"
description: |-
  The vm_disk resource allows to create a disk and attach it to a VM. The disk resource permit to create Internal or External disks. Internal create non-detachable disks and External create detachable disks.
---

# cloudavenue_vm_disk (Resource)

The `vm_disk` resource allows to create a disk and attach it to a VM. The disk resource permit to create Internal or External disks. Internal create non-detachable disks and External create detachable disks.

## Examples

### Internal Disk

```terraform
resource "cloudavenue_vm_disk" "example-internal" {
	vapp_id = cloudavenue_vapp.example.id
	name = "disk-example-internal"
	bus_type = "NVME"
	size_in_mb = 104800
	is_detachable = false
	vm_id = cloudavenue_vm.example.id
}
```

### External Disk

**External disk detached from VM**

```terraform
resource "cloudavenue_vm_disk" "example-detachable" {
	vapp_id = cloudavenue_vapp.example.id
	name = "disk-example"
	bus_type = "NVME"
	size_in_mb = 104800
	is_detachable = true
}
```

**External disk attached to VM**

```terraform
resource "cloudavenue_vm_disk" "example-detachable" {
	vapp_id = cloudavenue_vapp.example.id
	name = "disk-example"
	bus_type = "NVME"
	size_in_mb = 104800
	is_detachable = true
	vm_id = cloudavenue_vm.example.id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `size_in_mb` (Number) The size of the disk in MB.

### Optional

- `bus_number` (Number) (ForceNew) The bus number of the disk controller. If the disk is attached to a VM and this attribute is not set, the disk will be attached to the first available bus. Value must be between 0 and 3.
- `bus_type` (String) (ForceNew) The type of disk controller. Value must be one of : `IDE`, `SATA`, `SCSI`, `NVME`. Value defaults to `SCSI`.
- `is_detachable` (Boolean) (ForceNew) If set to true, the disk could be detached from the VM. If set to false, the disk canot detached to the VM. Value defaults to `false`.
- `name` (String) The name of the disk. If is_detachable attribute is set and the value is one of `true`, this attribute is REQUIRED. If is_detachable attribute is set and the value is one of `false`, this attribute is NULL.
- `storage_profile` (String) The name of the storage profile. If not set, the default storage profile will be used. Value must be one of : `silver`, `silver_r1`, `silver_r2`, `gold`, `gold_r1`, `gold_r2`, `gold_hm`, `platinum3k`, `platinum3k_r1`, `platinum3k_r2`, `platinum3k_hm`, `platinum7k`, `platinum7k_r1`, `platinum7k_r2`, `platinum7k_hm`.
- `unit_number` (Number) (ForceNew) The unit number of the disk controller. If the disk is attached to a VM and this attribute is not set, the disk will be attached to the first available unit. Value must be between 0 and 15.
- `vapp_id` (String) (ForceNew) ID of the vApp. Ensure that one and only one attribute from this collection is set : `vapp_name`, `vapp_id`.
- `vapp_name` (String) (ForceNew) Name of the vApp. Ensure that one and only one attribute from this collection is set : `vapp_id`, `vapp_name`.
- `vdc` (String) (ForceNew) The name of vDC to use, optional if defined at provider level.
- `vm_id` (String) The ID of the VM where the disk will be attached. Ensure that one and only one attribute from this collection is set : `vm_name`, `vm_id`.
- `vm_name` (String) The name of the VM where the disk will be attached. Ensure that one and only one attribute from this collection is set : `vm_id`, `vm_name`.

### Read-Only

- `id` (String) The ID of the Disk.

