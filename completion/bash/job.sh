
_dcos_job() {
    local i command
    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--config-schema"
    "--help"
    "--info"
    "--version"
    )

    local commands=(
    "add"
    "remove"
    "show"
    "update"
    "kill"
    "run"
    "list"
    "schedule"
    "show"
    "history"
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

_dcos_job_add() {
    # job add has no flags or subcommands
    return
}

_dcos_job_remove() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--stop-current-job-runs"
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


_dcos_job_update() {
    return
}

_dcos_job_kill() {
    local i command
    
    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--all"
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

_dcos_job_run() {
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

_dcos_job_list() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
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

_dcos_job_schedule() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local commands=(
    "add"
    "show"
    "remove"
    "update"
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

_dcos_job_schedule_add() {
    return
}

_dcos_job_schedule_show() {
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

_dcos_job_schedule_remove() {
    return
}

_dcos_job_schedule_update() {
    return
}

_dcos_job_show() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local commands=(
    "runs"
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

_dcos_job_show_runs() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--run-id="
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

_dcos_job_queue() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
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

_dcos_job_history() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--json"
    "--quiet"
    "--failures"
    "--last"
    )

    if [ -z "$command" ]; then
        case "$cur" in
            --*)
                __dcos_handle_compreply "$flags[@]}"
                ;;
            *)
                ;;
        esac
        return
    fi
}
