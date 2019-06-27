
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
    "download"
    "exec"
    "list"
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
                __dcos_handle_compreply "${flags[@]}"
                ;;
            *)
                __dcos_complete_task_ids
                ;;
        esac
        return
    fi
}

_dcos_task_download() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--target-dir="
    )

    if [ -z "$command" ]; then
        case "$cur" in
            --*)
                __dcos_handle_compreply "${flags[@]}"
                ;;
            *)
                __dcos_complete_task_ids
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
    "--user"
    )

    if [ -z "$command" ]; then
        case "$cur" in
            --*)
                __dcos_handle_compreply "${flags[@]}"
                ;;
            *)
                __dcos_complete_task_ids
                ;;
        esac
        return
    fi
}

_dcos_task_list() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--all"
    "--completed"
    "--json"
    "--quiet"
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
                __dcos_complete_task_ids
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
                __dcos_complete_task_ids
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
                __dcos_complete_task_ids
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
                __dcos_complete_task_ids
                ;;
        esac
        return
    fi

}

__dcos_complete_task_ids() {
    while IFS=$'\n' read -r line; do task_ids+=("$line"); done < <(dcos task list --quiet 2> /dev/null)
    __dcos_handle_compreply "${task_ids[@]}"
}

