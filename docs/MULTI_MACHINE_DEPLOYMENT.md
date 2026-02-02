# Multi-Machine/Client Deployment Guide

## 🚨 Problem Statement

Previously, running `make destroy-studio` from Machine B (Mac Pro) would delete the Pulumi state for Machine A's (Mac Studio) cluster, even though:
- The actual Kind cluster on Machine A would still be running
- Pulumi would lose track of what was deployed
- The cluster would become "orphaned" - running but unmanaged

## ✅ Solution: Machine-Specific Stack Naming

### How It Works

1. **Automatic Machine Detection**: The system detects the hostname and creates a machine ID
   - `studio` machine → machine ID: `studio`
   - `pro` machine → machine ID: `pro`
   - Other machines → machine ID: sanitized hostname

2. **Stack Naming Format**: `org/env-machine`
   - Old: `bvlucena/studio` (shared across all machines ❌)
   - New: `bvlucena/studio-studio` (Mac Studio machine ✅)
   - New: `bvlucena/studio-pro` (Mac Pro machine ✅)
   - New: `bvlucena/studio-client1` (Client 1 machine ✅)

3. **Safety Checks in Destroy Command**:
   - ✅ Verifies Kind cluster exists **locally** before allowing destruction
   - ✅ Shows machine hostname and ID in all operations
   - ✅ Requires explicit confirmation with machine name
   - ✅ Prevents accidental cross-machine destruction

## 📋 Usage Examples

### Deploying on Multiple Machines

```bash
# On Mac Studio
make up ENV=studio
# Creates: bvlucena/studio-studio

# On Mac Pro
make up ENV=studio  
# Creates: bvlucena/studio-pro

# On Client Machine
make up ENV=studio
# Creates: bvlucena/studio-<client-hostname>
```

### Listing All Stacks

```bash
make list-stacks
# Shows all stacks across all machines/clients
```

### Destroying (Protected)

```bash
# On Mac Studio - this will work
make destroy ENV=studio
# ✅ Cluster exists locally, destruction proceeds

# On Mac Pro - this will FAIL
make destroy ENV=studio
# ❌ ERROR: Cluster 'studio' does NOT exist on this machine
#    This prevents accidental destruction of clusters on other machines.
```

### Checking Environment Info

```bash
make env
# Shows:
#   - Current environment
#   - Pulumi stack name (with machine ID)
#   - Machine ID and hostname
#   - Local Kind clusters
```

## 🔧 Technical Details

### Machine Detection Logic

The system uses hostname detection with fallbacks:

```makefile
MACHINE_HOSTNAME := $(shell hostname | tr '[:upper:]' '[:lower:]' | sed 's/\.local$$//' | sed 's/[^a-z0-9-]/-/g')

# Detect machine type from hostname
ifeq (studio in hostname)
  MACHINE_ID := studio
else ifeq (pro in hostname)
  MACHINE_ID := pro
else
  MACHINE_ID := sanitized-hostname
endif
```

### Safety Check Implementation

The `destroy` command includes:

1. **Cluster Existence Check**:
   ```bash
   if ! kind get clusters | grep -qw "$(CLUSTER_NAME)"; then
     echo "❌ ERROR: Cluster does NOT exist on this machine"
     exit 1
   fi
   ```

2. **Cluster Accessibility Check**:
   ```bash
   kubectl cluster-info --context kind-$(CLUSTER_NAME)
   ```

3. **Stack Validation**:
   ```bash
   pulumi stack select $(PULUMI_STACK)
   ```

## 🎯 Benefits

1. **Client Deployments**: Each client can deploy the same environment without conflicts
2. **Multi-Machine Safety**: Can't accidentally destroy clusters on other machines
3. **State Isolation**: Each machine maintains its own Pulumi state
4. **Clear Identification**: Stack names clearly show which machine they belong to

## 🔄 Migration Notes

If you have existing stacks with the old naming:

1. **Old stacks will continue to work** but are not protected
2. **New deployments** automatically use machine-specific naming
3. **To migrate**: Destroy old stack and recreate (will use new naming)

## 📚 Related Commands

- `make env` - Show current environment and machine info
- `make list-stacks` - List all stacks across machines
- `make destroy ENV=<name>` - Destroy with safety checks
- `make up ENV=<name>` - Deploy with machine-specific stack

---

**Last Updated**: 2025-01-27  
**Status**: ✅ Implemented and Protected
