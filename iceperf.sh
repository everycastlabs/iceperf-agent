#!/bin/bash

# Determine the architecture
arch=$(uname -m)

# Set the appropriate file names and URLs based on the architecture
case "$arch" in
    x86_64)
        asset_name="iceperf-agent-linux-amd64.tar.gz"
        ;;
    arm64)
        asset_name="iceperf-agent-linux-arm64.tar.gz"
        ;;
    *)
        echo "Unsupported architecture: $arch"
        exit 1
        ;;
esac

asset_url="https://github.com/nimbleape/iceperf-agent/releases/latest/download/$asset_name"
md5_url="$asset_url.md5"
local_file="$asset_name"
binary_name="iceperf-agent"

# Clean up any existing files
rm -f "$local_file" || true

# Download the remote MD5 hash
remote_md5=$(wget -qO- "$md5_url")

# Extract the MD5 hash from the downloaded file
current_md5=$(cat current.md5)

# Compare the hashes

echo "Current MD5 is $current_md5"
echo "Remote MD5 is $remote_md5"

if [ "$current_md5" == "$remote_md5" ] && [ -f "$binary_name" ];then
    echo "The file is already up-to-date."
else
    echo "The file is outdated or does not exist. Downloading the new version."
    rm "$binary_name" || true
    rm current.md5 || true
    wget -O "$local_file" "$asset_url"
    wget -O current.md5 "$md5_url"
    echo "Downloaded new client and new current.md5"

    # Extract the downloaded tar.gz file
    tar -xzvf "$local_file"

    # Make the binary executable
    chmod +x "$binary_name"
fi

echo "Calling $binary_name with API Key"
# Run the binary with the api key
./"$binary_name" --api-key="your-api-key"

# Clean up
rm -f "$local_file"