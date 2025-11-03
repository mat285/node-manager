#!/bin/sh

set -e

VERSION=$1
if [ -z "$VERSION" ]; then
    VERSION=$(curl -sL https://api.github.com/repos/mat285/node-manager/releases/latest | jq -r .name)
    echo "No version specified, using latest release: ${VERSION}"
fi

which kubectl >/dev/null 2>&1 || {
    echo "kubectl is not installed. Please install kubectl first."
    exit 1
}

OS=$(uname -s | dd conv=lcase 2>/dev/null)
ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
    ARCH="amd64"
elif [ "$ARCH" = "aarch64" ]; then
    ARCH="arm64"
fi

if [ "$ARCH" != "amd64" ] && [ "$ARCH" != "arm64" ]; then
    echo "Unsupported architecture: ${ARCH}"
    exit 1
fi

if [ "$OS" != "linux" ] && [ "$OS" != "darwin" ]; then
    echo "Unsupported OS: ${OS}"
    exit 1
fi

echo "Detected OS: ${OS}, Architecture: ${ARCH}"

sudo systemctl stop node-manager.service || true
sudo systemctl disable node-manager.service || true
echo "Installing node-manager server version ${VERSION}..."

echo "Downloading https://github.com/mat285/node-manager/releases/download/${VERSION}/node-manager_${OS}_${ARCH}"
curl --fail-with-body -Lo node-manager https://github.com/mat285/node-manager/releases/download/${VERSION}/node-manager_${OS}_${ARCH}
sudo mv node-manager /bin/node-manager
sudo chown root:root /bin/node-manager
sudo chmod a+x /bin/node-manager

echo "Downloading node-manager.service https://github.com/mat285/node-manager/releases/download/${VERSION}/node-manager.service"
curl --fail-with-body -Lo node-manager.service https://github.com/mat285/node-manager/releases/download/${VERSION}/node-manager.service
sudo mv node-manager.service /etc/systemd/system/node-manager.service
sudo chmod 644 /etc/systemd/system/node-manager.service
sudo chown root:root /etc/systemd/system/node-manager.service

echo "Downloading example.yml https://github.com/mat285/node-manager/releases/download/${VERSION}/example.yml"
curl --fail-with-body -Lo example.yml https://github.com/mat285/node-manager/releases/download/${VERSION}/example.yml
sudo mkdir -p /etc/node-manager
sudo mv example.yml /etc/node-manager/example.yml
sudo chown root:root /etc/node-manager/example.yml

sudo systemctl daemon-reload
sudo systemctl enable node-manager.service
sudo systemctl start node-manager.service
echo "node-manager server installed and started successfully."
