#!/bin/bash
function usage() {
  cat << _EOF_
Usage: ${0} apikey [options]
  apikey            Your API key
  [options]         Any options to pass to the iceperf-agent

Example: ${0} your-api-key --timer

_EOF_
}

#-- check arguments and environment
if [ "$#" -lt 1 ]; then
  echo "Expected at least 1 argument, got $#"
  usage
  exit 2
fi

APIKEY=$1
shift

# Determine the architecture
arch=$(uname -m)
os=$(uname -s | tr '[:upper:]' '[:lower:]')
# Set the appropriate file names and URLs based on the architecture
case "$arch" in
    x86_64)
        asset_name="iceperf-agent-$os-amd64.tar.gz"
        ;;
    arm64)
        asset_name="iceperf-agent-$os-arm64.tar.gz"
        ;;
    *)
        echo "Unsupported architecture: $arch"
        exit 1
        ;;
esac

# Check if wget exists
if ! command -v wget >/dev/null 2>&1; then
  echo "wget is not installed. Please install it."
  exit 1
fi

asset_url="https://github.com/nimbleape/iceperf-agent/releases/latest/download/$asset_name"
md5_url="$asset_url.md5"
local_file="$asset_name"
binary_name="iceperf-agent"

# Clean up any existing files
rm -f "$local_file" || true

# Download the remote MD5 hash
remote_md5=$(wget -qO- "$md5_url")

# Extract the MD5 hash from the downloaded file
if [ -f current.md5 ]; then
  current_md5=$(cat current.md5)
else
  current_md5=""
fi

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

touch current.md5

echo "Calling $binary_name with API Key"
# Run the binary with the api key
./"$binary_name" --api-key="${APIKEY}" "$@"

# Clean up
rm -f "$local_file"