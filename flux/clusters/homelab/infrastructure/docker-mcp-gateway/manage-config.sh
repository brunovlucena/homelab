#!/bin/bash
# MCP Gateway Configuration Management Script

set -e

NAMESPACE="mcp-gateway"
DEPLOYMENT="mcp-gateway"
CONFIG_DIR="/home/mcp/.docker/mcp"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

usage() {
    cat << EOF
MCP Gateway Configuration Management

Usage: $0 <command> [options]

Commands:
    backup              Backup current configuration
    restore <dir>       Restore configuration from backup directory
    view <file>         View a configuration file (docker-mcp.yaml, registry.yaml, config.yaml, tools.yaml)
    edit <file>         Edit a configuration file
    export              Export all configuration files to current directory
    import <dir>        Import configuration files from directory
    list-servers        List all configured servers
    enable <server>     Enable a server
    disable <server>    Disable a server
    restart             Restart the gateway deployment

Examples:
    $0 backup
    $0 view config.yaml
    $0 edit docker-mcp.yaml
    $0 enable brave-search
    $0 export
    $0 import ./my-configs/
    $0 restart

EOF
    exit 1
}

get_pod() {
    kubectl get pod -n $NAMESPACE -l app.kubernetes.io/name=$DEPLOYMENT -o jsonpath='{.items[0].metadata.name}' 2>/dev/null
}

wait_for_pod() {
    echo -e "${YELLOW}Waiting for pod to be ready...${NC}"
    kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=$DEPLOYMENT -n $NAMESPACE --timeout=60s
}

backup_config() {
    POD=$(get_pod)
    if [ -z "$POD" ]; then
        echo -e "${RED}Error: No pod found${NC}"
        exit 1
    fi

    BACKUP_DIR="backups/mcp-gateway-$(date +%Y%m%d-%H%M%S)"
    mkdir -p $BACKUP_DIR

    echo -e "${GREEN}Backing up configuration to $BACKUP_DIR${NC}"
    kubectl cp $NAMESPACE/$POD:$CONFIG_DIR/ $BACKUP_DIR/

    echo -e "${GREEN}✓ Backup completed: $BACKUP_DIR${NC}"
    ls -lh $BACKUP_DIR/
}

restore_config() {
    if [ -z "$1" ]; then
        echo -e "${RED}Error: Backup directory required${NC}"
        usage
    fi

    if [ ! -d "$1" ]; then
        echo -e "${RED}Error: Directory $1 not found${NC}"
        exit 1
    fi

    POD=$(get_pod)
    if [ -z "$POD" ]; then
        echo -e "${RED}Error: No pod found${NC}"
        exit 1
    fi

    echo -e "${YELLOW}Restoring configuration from $1${NC}"
    kubectl cp $1/ $NAMESPACE/$POD:$CONFIG_DIR/

    echo -e "${GREEN}✓ Configuration restored${NC}"
    echo -e "${YELLOW}Restarting deployment...${NC}"
    kubectl rollout restart deployment/$DEPLOYMENT -n $NAMESPACE
}

view_file() {
    if [ -z "$1" ]; then
        echo -e "${RED}Error: Filename required${NC}"
        usage
    fi

    POD=$(get_pod)
    if [ -z "$POD" ]; then
        echo -e "${RED}Error: No pod found${NC}"
        exit 1
    fi

    echo -e "${GREEN}Viewing $CONFIG_DIR/$1${NC}"
    kubectl exec -n $NAMESPACE $POD -- cat $CONFIG_DIR/$1
}

edit_file() {
    if [ -z "$1" ]; then
        echo -e "${RED}Error: Filename required${NC}"
        usage
    fi

    POD=$(get_pod)
    if [ -z "$POD" ]; then
        echo -e "${RED}Error: No pod found${NC}"
        exit 1
    fi

    TEMP_FILE=$(mktemp)
    echo -e "${GREEN}Downloading $1...${NC}"
    kubectl exec -n $NAMESPACE $POD -- cat $CONFIG_DIR/$1 > $TEMP_FILE

    ${EDITOR:-vi} $TEMP_FILE

    echo -e "${GREEN}Uploading modified $1...${NC}"
    kubectl exec -i -n $NAMESPACE $POD -- sh -c "cat > $CONFIG_DIR/$1" < $TEMP_FILE

    rm $TEMP_FILE
    echo -e "${GREEN}✓ File updated${NC}"
    echo -e "${YELLOW}Restart deployment to apply changes: $0 restart${NC}"
}

export_config() {
    POD=$(get_pod)
    if [ -z "$POD" ]; then
        echo -e "${RED}Error: No pod found${NC}"
        exit 1
    fi

    EXPORT_DIR="mcp-gateway-config-$(date +%Y%m%d-%H%M%S)"
    mkdir -p $EXPORT_DIR

    echo -e "${GREEN}Exporting configuration to $EXPORT_DIR${NC}"
    kubectl cp $NAMESPACE/$POD:$CONFIG_DIR/ $EXPORT_DIR/

    echo -e "${GREEN}✓ Configuration exported: $EXPORT_DIR${NC}"
    ls -lh $EXPORT_DIR/
}

import_config() {
    if [ -z "$1" ]; then
        echo -e "${RED}Error: Import directory required${NC}"
        usage
    fi

    if [ ! -d "$1" ]; then
        echo -e "${RED}Error: Directory $1 not found${NC}"
        exit 1
    fi

    POD=$(get_pod)
    if [ -z "$POD" ]; then
        echo -e "${RED}Error: No pod found${NC}"
        exit 1
    fi

    echo -e "${YELLOW}Importing configuration from $1${NC}"
    
    for file in docker-mcp.yaml registry.yaml config.yaml tools.yaml; do
        if [ -f "$1/$file" ]; then
            echo -e "${GREEN}Importing $file...${NC}"
            kubectl cp $1/$file $NAMESPACE/$POD:$CONFIG_DIR/$file
        fi
    done

    echo -e "${GREEN}✓ Configuration imported${NC}"
    echo -e "${YELLOW}Restart deployment to apply changes: $0 restart${NC}"
}

list_servers() {
    POD=$(get_pod)
    if [ -z "$POD" ]; then
        echo -e "${RED}Error: No pod found${NC}"
        exit 1
    fi

    echo -e "${GREEN}Configured Servers:${NC}"
    kubectl exec -n $NAMESPACE $POD -- cat $CONFIG_DIR/docker-mcp.yaml | grep -A 3 "servers:" || echo "No servers configured"
    
    echo -e "\n${GREEN}Enabled Servers:${NC}"
    kubectl exec -n $NAMESPACE $POD -- cat $CONFIG_DIR/registry.yaml
}

restart_deployment() {
    echo -e "${YELLOW}Restarting $DEPLOYMENT...${NC}"
    kubectl rollout restart deployment/$DEPLOYMENT -n $NAMESPACE
    kubectl rollout status deployment/$DEPLOYMENT -n $NAMESPACE
    echo -e "${GREEN}✓ Deployment restarted${NC}"
}

# Main command dispatcher
case "${1:-}" in
    backup)
        backup_config
        ;;
    restore)
        restore_config "$2"
        ;;
    view)
        view_file "$2"
        ;;
    edit)
        edit_file "$2"
        ;;
    export)
        export_config
        ;;
    import)
        import_config "$2"
        ;;
    list-servers)
        list_servers
        ;;
    restart)
        restart_deployment
        ;;
    *)
        usage
        ;;
esac

