
_dcos_node() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--help"
    "--json"
    "--field="
    )

    local commands=(
    "decommision"
    "diagnostics"
    "dns"
    "list"
    "list-components"
    "log"
    "metrics"
    "ssh"
    )

    if [ -z "$command" ]; then
        case "$cur" in
            --*)
                __dcos_handle_compreply "${flags[@]}"
                ;;
            *)
                __dcos_handle_compreply "${commands[@]}"
                ;;
        esac
        return
    fi

    __dcos_handle_subcommand
}

_dcos_node_decommision() {
    return
}

_dcos_node_diagnostics() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--json"
    )

    local commands=(
    "cancel"
    "create"
    "delete"
    "download"
    "list"
    "status"
    )

    if [ -z "$command" ]; then
        case "$cur" in
            --*)
                __dcos_handle_compreply "${flags[@]}"
                ;;
            *)
                __dcos_handle_compreply "${commands[@]}"
                ;;
        esac
        return
    fi

    __dcos_handle_subcommand
}

_dcos_node_diagnostics_cancel() {
    return
}

_dcos_node_diagnostics_create() {
    return
}

_dcos_node_diagnostics_delete() {
    return
}

_dcos_node_diagnostics_download() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--location="
    )

    if [ -z "$command" ]; then
        case "$cur" in
            --*)
                __dcos_handle_compreply "${flags[@]}"
                ;;
            *)
                ;;
        esac
        return
    fi
}

_dcos_node_diagnostics_list() {
    return
}

_dcos_node_diagnostics_status() {
    return
}

_dcos_node_dns() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--json"
    )

    if [ -z "$command" ]; then
        case "$cur" in
            --*)
                __dcos_handle_compreply "${flags[@]}"
                ;;
            *)
                ;;
        esac
        return
    fi
}

_dcos_node_list() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--help"
    "--json"
    "--field="
    )

    if [ -z "$command" ]; then
        case "$cur" in
            --*)
                __dcos_handle_compreply "${flags[@]}"
                ;;
            --)
                ;;
        esac
        return
    fi
}

_dcos_node_list_components() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--leader"
    "--mesos-id="
    "--json"
    )

    if [ -z "$command" ]; then
        case "$cur" in
            --*)
                __dcos_handle_compreply "${flags[@]}"
                ;;
            --)
                ;;
        esac
        return
    fi
}

_dcos_node_log() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--follow"
    "--lines="
    "--leader"
    "--mesos-id="
    "--component="
    "--filter="
    )

    if [ -z "$command" ]; then
        case "$cur" in
            --*)
                __dcos_handle_compreply "${flags[@]}"
                ;;
            *)
                ;;
        esac
        return
    fi
}

_dcos_node_metrics() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local commands=(
    "details"
    "summary"
    )

    if [ -z "$command" ]; then
        case "$cur" in
            --*)
                ;;
            *)
                __dcos_handle_compreply "${commands[@]}"
                ;;
        esac
        return
    fi

    __dcos_handle_subcommand
}

_dcos_node_metrics_details() {

    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--json"
    )

    if [ -z "$command" ]; then
        case "$cur" in
            --*)
                __dcos_handle_compreply "${flags[@]}"
                ;;
            *)
                ;;
        esac
        return
    fi
}

_dcos_node_metrics_summary() {

    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--json"
    )

    if [ -z "$command" ]; then
        case "$cur" in
            --*)
                __dcos_handle_compreply "${flags[@]}"
                ;;
            *)
                ;;
        esac
        return
    fi
}

_dcos_node_ssh() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--leader"
    "--mesos-id="
    "--private-ip="
    "--config-file="
    "--user="
    "--master-proxy"
    # --option takes in arguments of the form OPT=VAL, not sure what do really
    # do here so I'll leave off the = at the end of --option
    "--option"
    "--proxy-ip="
    )

    if [ -z "$command" ]; then
        case "$cur" in
            --*)
                __dcos_handle_compreply "${flags[@]}"
                ;;
            *)
                ;;
        esac
        return
    fi
}
