# kubectl-image

[![Release](https://github.com/reedchan7/kubectl-image/actions/workflows/release.yaml/badge.svg)](https://github.com/reedchan7/kubectl-image/actions/workflows/release.yaml)

A simple, fast, and safe `kubectl` plugin to manage container images in your Kubernetes resources.

---

## Features

-   **Get Images**: Quickly retrieve the image of a `deployment` or `pod`.
-   **Set Images**: Update the image of a `deployment` safely.
    -   Updates only the first container by default to prevent accidental changes.
    -   Supports updating a specific container with the `--container` flag.
-   **Tag-focused**:
    -   Get just the image tag with `get --tag`.
    -   Set just the image tag with `set --tag`.
-   **Clean & Simple**: A focused CLI that does one thing well. No unnecessary flags or complexity.

## Usage

The plugin follows the standard `kubectl` command structure.

### Get Image

Get the full image name of the first container in a resource.

```sh
# Get image of a deployment
$ kubectl image get deployment my-app
busybox:latest

# Get image of a pod
$ kubectl image get pod my-pod-12345
busybox:latest
```

To get only the **tag** of the image:

```sh
$ kubectl image get deployment my-app --tag
v1.2.3
```

### Set Image

Update the image of the first container in a deployment.

```sh
# Set a full new image
$ kubectl image set deployment my-app busybox:1.36
Updating container my-app image from busybox:latest to busybox:1.36
deployment.apps/my-app image updated

# Update only the tag, keeping the base image name
$ kubectl image set deployment my-app --tag 1.36.1
Updating container my-app image from busybox:1.36 to busybox:1.36.1
deployment.apps/my-app image updated

# Update a specific container within the deployment
$ kubectl image set deploy my-app --tag v2.0.2 --container sidecar
```

## Installation

There are several ways to install `kubectl-image`.

### Krew (TODO)

### From GitHub Releases

You can install `kubectl-image` with a single command using our installation script. It will automatically download the correct binary for your system from the latest GitHub Release.

```sh
curl -fsSL https://raw.githubusercontent.com/reedchan7/kubectl-image/main/install.sh | sh
```

### With `go install`

If you have a Go environment set up, you can install directly from the source:
```sh
go install github.com/reedchan7/kubectl-image/src/cmd/kubectl-image@latest
```
*Note: Make sure your `$GOPATH/bin` or `$HOME/go/bin` directory is in your system's `$PATH`.*

### From Source (for Developers)

1.  Clone the repository:
    ```sh
    git clone https://github.com/reedchan7/kubectl-image.git
    cd kubectl-image
    ```

2.  Build and install the plugin using `make`:
    ```sh
    make install
    ```
    This command will compile the source and copy the binary to `/usr/local/bin` using `sudo`.

3.  Verify the installation:
    ```sh
    kubectl plugin list
    ```
    You should see `image` listed as a valid plugin.

## Development

-   **Build**: `make build`
-   **Test**: `make test`
-   **Clean**: `make clean`

## Uninstall

To remove the plugin from your system:

```sh
make uninstall
```

## License

This project is open-source and available under the [Apache-2.0](LICENSE).
