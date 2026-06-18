## Module availability

The catalog (`dce commands` / `dce search`) lists every module `dce` was built with, **regardless of which modules are actually installed on the target host**. A command appearing in the catalog does not guarantee its module is deployed.

A `404` / route-not-found at the **module path level** (the API route itself does not exist, not a specific object) is ambiguous: the module may not be installed, or the path/version may be wrong. On such a `404`, **actively confirm before concluding** by running `dce global-management about list-g-product-versions -o json` (global-management is the always-present base module, so this is always available). Then:

- **Module not in the list** → it is not installed on this host. Do not retry, do not try sibling commands in the same module, and do not tell the user the capability exists — state plainly that the module is not installed.
- **Module in the list** → it is installed, so the `404` is not a missing module. Re-check the exact path and params with `dce commands show <path...> --json` (likely a wrong path/version or a resource-level not-found); do not report the module as missing.
- **Resource-level `404`** (a specific ID/name on an otherwise-working module) means **that object** doesn't exist, not that the module is missing — handle that normally.
