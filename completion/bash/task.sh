
_dcos_task() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--help"
    "--info"
    "--version"
    "--all"
    "--completed"
    "--json"
    )

    local commands=(
    "attach"
    "exec"
    "log"
    "ls"
    "metrics"
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

_dcos_task_attach() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--no-stdin"
    )

    if [ -z "$command" ]; then
        case "$cur" in
            --*)
                __dcos_handle_compreply "${flgs[@]}"
                ;;
            *)
                ;;
        esac
        return
    fi
}

_dcos_task_exec() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--interactive"
    "--tty"
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

_dcos_task_log() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--all"
    "--completed"
    "--follow"
    "--lines="
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

_dcos_task_ls() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--all"
    "--completed"
    "--long"
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

_dcos_task_metrics() {
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

_dcos_task_metrics_details() {
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

_dcos_task_metrics_summary() {
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

