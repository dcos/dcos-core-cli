
_dcos_diagnostics() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--help"
    )

    local commands=(
    "create"
    "delete"
    "download"
    "list"
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

_dcos_diagnostics_create() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--help"
    )

    if [ -z "$command" ]; then
        case "$cur" in
            --*)
                __dcos_handle_compreply "${flags[@]}"
                ;;
        esac
        return
    fi
}

_dcos_diagnostics_delete() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--help"
    )

    if [ -z "$command" ]; then
        case "$cur" in
            --*)
                __dcos_handle_compreply "${flags[@]}"
                ;;
            *)
                __dcos_complete_bundle_ids
                ;;
        esac
        return
    fi
}

_dcos_diagnostics_download() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--help"
    "--output"
    )

    # the normal `if` here omitted to make the output file completion work
    case "$cur" in
        --*)
            __dcos_handle_compreply "${flags[@]}"
            ;;
        *)
            if [ "$prev" == "--output" ]; then
                # if the last word was output then we return which causes
                # bash to use the default file completion
                return
            fi
            __dcos_complete_bundle_ids
            ;;
    esac
}

_dcos_diagnostics_list() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--help"
    "--json"
    )

    if [ -z "$command" ]; then
        case "$cur" in
            --*)
                __dcos_handle_compreply "${flags[@]}"
                ;;
        esac
        return
    fi
}

__dcos_complete_bundle_ids() {
    while IFS=$'\n' read -r line; do bundle_ids+=("$line"); done < <(dcos diagnostics list -q 2> /dev/null)
    __dcos_handle_compreply "${bundle_ids[@]}"
}
