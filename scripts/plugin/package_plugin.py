
import shutil
import os
from os import path

from distutils import dir_util, file_util

plugin_toml_template = '''
schema_version = 1
name = "dcos-core-cli"

[[commands]]
name = "job"
path = "bin/dcos_py{0}"
description = "Deploy and manage jobs in DC/OS"

[[commands]]
name = "marathon"
path = "bin/dcos_py{0}"
description = "Deploy and manage applications to DC/OS"

[[commands]]
name = "node"
path = "bin/dcos_py{0}"
description = "View DC/OS node information"

[[commands]]
name = "package"
path = "bin/dcos_py{0}"
description = "Install and manage DC/OS software packages"

[[commands]]
name = "service"
path = "bin/dcos_py{0}"
description = "Manage DC/OS services"

[[commands]]
name = "task"
path = "bin/dcos_py{0}"
description = "Manage DC/OS tasks"
'''

# Path to the root of the repo
root_path = path.join(
    path.dirname(path.realpath(__file__)), "..", ".."
)

def create_plugin_toml(filepath: str, platform: str):
    plugin_path = path.join(build_path, platform, "plugin")
    bin_extension = '.exe' if platform == 'windows' else ''

    with open(filepath, encoding='utf-8', mode='w') as file:
        file.write(plugin_toml_template.format(bin_extension))


def package_completions(plugin_path: str):
    completions_path = path.join(root_path, "completion")

    dest_path = path.join(
        plugin_path, "completion"
    )
    # only copy completion dir if it's there
    if path.exists(completions_path):
        dir_util.copy_tree(completions_path, dest_path)


def package_binaries(plugin_path: str, platform: str):
    bin_extension = ".exe" if platform == "windows" else ""

    go_bin = path.join(plugin_path, "..", "dcos{}".format(bin_extension))
    python_bin = path.join(root_path, "python", "lib", "dcoscli", "dist", "dcos{}".format(bin_extension))

    dest = path.join(plugin_path, "bin")
    dir_util.mkpath(dest)

    # As we aren't using the Go CLI piece yet, this shouldn't be moved into the folder
    # file_util.copy_file(go_bin, path.join(dest, "dcos"))
    file_util.copy_file(python_bin, path.join(dest, "dcos_py{}".format(bin_extension)))


def package_plugin(plugin_path: str, platform: str):
    if not path.exists(plugin_path):
        os.makedirs(plugin_path)

    toml_path = path.join(plugin_path, "plugin.toml")
    create_plugin_toml(toml_path, platform)

    package_completions(plugin_path)

    package_binaries(plugin_path, platform)

    target_filepath = path.join(build_path, platform, "dcos-core-cli")
    shutil.make_archive(
        target_filepath,
        'zip',
        plugin_path
    )


if __name__ == '__main__':
    build_path = path.join(root_path, "build")

    # Because pyinstaller does not allow cross compilation, when running this
    # as __main__, it only packages a plugin for the current platform. It
    # assumes that pyinstaller has already created the binary prior to this
    # being run.
    platform = os.uname().sysname.lower()
    platform_build_path = path.join(build_path, platform)
    plugin_path = path.join(platform_build_path, "plugin")

    if not path.exists(platform_build_path):
        os.mkdir(platform_build_path)

    package_plugin(plugin_path, platform)

