
_dcos_marathon() {
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
    "about"
    "app"
    "deployment"
    "group"
    "leader"
    "ping"
    "plugin"
    "pod"
    "debug"
    "task"
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

_dcos_marathon_about() {
    return
}

_dcos_marathon_app() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local commands=(
    "add"
    "list"
    "remove"
    "restart"
    "show"
    "start"
    "stop"
    "kill"
    "update"
    "version"
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

_dcos_marathon_app_add() {
    return
}

_dcos_marathon_app_list() {
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

_dcos_marathon_app_remove() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--force"
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

_dcos_marathon_app_restart() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--force"
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

_dcos_marathon_app_show() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--app-version="
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

_dcos_marathon_app_start() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--force"
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


_dcos_marathon_app_stop() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--force"
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

_dcos_marathon_app_kill() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--scale"
    "--host="
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

_dcos_marathon_app_update() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--force"
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

_dcos_marathon_app_version() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local commands=(
    "list"
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

_dcos_marathon_app_version_list() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--max-count="
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

_dcos_marathon_deployment() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local commands=(
    "list"
    "rollback"
    "stop"
    "watch"
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

_dcos_marathon_deployment_list() {
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

_dcos_marathon_deployment_rollback() {
    return
}

_dcos_marathon_deployment_stop() {
    return
}

_dcos_marathon_deployment_watch() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--max-count="
    "--interval="
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

_dcos_marathon_group() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local commands=(
    "add"
    "list"
    "scale"
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

_dcos_marathon_group_add() {
    return
}

_dcos_marathon_group_list() {
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

_dcos_marathon_group_scale() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--force"
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

_dcos_marathon_group_show() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--group-version="
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

_dcos_marathon_group_remove() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--force"
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

_dcos_marathon_group_update() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--force"
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

_dcos_marathon_leader() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local commands=(
    "delete"
    "show"
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

_dcos_marathon_leader_delete() {
    return
}

_dcos_marathon_leader_show() {
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

_dcos_marathon_ping() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--once"
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

_dcos_marathon_plugin() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local commands=(
    "list"
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

_dcos_marathon_plugin_list() {
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

_dcos_marathon_pod() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local commands=(
    "add"
    "kill"
    "list"
    "remove"
    "show"
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
    fi

    __dcos_handle_subcommand
}

_dcos_marathon_pod_add() {
    return
}

_dcos_marathon_pod_kill() {
    return
}

_dcos_marathon_pod_list() {
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

_dcos_marathon_pod_remove() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--force"
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

_dcos_marathon_pod_show() {
    return
}

_dcos_marathon_pod_update() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--force"
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

_dcos_marathon_debug() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local commands=(
    "list"
    "summary"
    "details"
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

_dcos_marathon_debug_list() {
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

_dcos_marathon_debug_summary() {
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

_dcos_marathon_debug_details() {
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

_dcos_marathon_task() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local commands=(
    "list"
    "stop"
    "kill"
    "show"
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

_dcos_marathon_task_list() {
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

_dcos_marathon_task_stop() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--wipe"
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

_dcos_marathon_task_kill() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--scale"
    "--wipe"
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

_dcos_marathon_task_show() {
    return
}
