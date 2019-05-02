
_dcos_service() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--help"
    "--info"
    "--version"
    "--completed"
    "--inactive"
    "--json"
    )

    local commands=(
    "log"
    "shutdown"
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

_dcos_service_log() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--follow"
    "--lines="
    "--ssh-config-file="
    )

    if [ -z "$command" ]; then
        case "$cur" in
            --*)
                __dcos_handle_compreply "${flags[@]}"
                ;;
            *)
                __dcos_complete_service_names
                ;;
        esac
        return
    fi
}

_dcos_service_shutdown() {
    __dcos_complete_service_names
}

__dcos_complete_service_names() {
    while IFS=$'\n' read -r line; do service_names+=("$line"); done < <(dcos service list -q 2> /dev/null)
    __dcos_handle_compreply "${service_names[@]}"
}
