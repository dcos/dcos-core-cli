
_dcos_package() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--config-schema"
    "--help"
    "--info"
    )

    local commands=(
    "describe"
    "install"
    "list"
    "repo"
    "search"
    "uninstall"
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

_dcos_package_describe() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--app"
    "--render"
    "--cli"
    "--config"
    "--options="
    "--package-version="
    "--package-versions"
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

_dcos_package_install() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--cli"
    "--global"
    "--app"
    "--package-version="
    "--options="
    "--yes"
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

_dcos_package_list() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--json"
    "--app-id="
    "--cli"
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

_dcos_package_repo() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local commands=(
    "add"
    "import"
    "list"
    "remove"
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

_dcos_package_repo_add() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--index="
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

_dcos_package_repo_import() {
    return
}

_dcos_package_repo_list() {
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

_dcos_package_repo_remove() {
    return
}

_dcos_package_search() {
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

_dcos_package_uninstall() {
    local i command

    if ! __dcos_default_command_parse; then
        return
    fi

    local flags=(
    "--cli"
    "--app"
    "--app-id="
    "--all"
    "--yes"
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

