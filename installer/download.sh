# This is the core installation part script for b4 Universal.

# Get latest release version from GitHub - ONLY returns version string
get_latest_version() {
    api_url="https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest"
    version=""

    # Try wget first, then curl
    if command_exists wget; then
        version=$(wget -qO- "$api_url" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command_exists curl; then
        version=$(curl -s "$api_url" 2>/dev/null | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        print_error "Neither wget nor curl found. Please install one of them."
        exit 1
    fi

    if [ -z "$version" ]; then
        print_error "Failed to fetch latest version"
        exit 1
    fi

    # ONLY output the result to stdout
    echo "$version"
}

# Verify checksum
verify_checksum() {
    file="$1"
    checksum_url="$2"

    checksum_file="${file}.sha256"

    # Try to download checksum file
    print_info "Downloading SHA256 checksum..."
    if command_exists wget; then
        if ! wget -q -O "$checksum_file" "$checksum_url" 2>/dev/null; then
            return 1
        fi
    elif command_exists curl; then
        if ! curl -s -L -o "$checksum_file" "$checksum_url" 2>/dev/null; then
            return 1
        fi
    else
        return 1
    fi

    # Check if checksum file was actually downloaded (not a 404 page)
    if [ ! -s "$checksum_file" ]; then
        rm -f "$checksum_file"
        return 1
    fi

    # Extract expected checksum (handle format: "checksum filename")
    expected_checksum=$(cat "$checksum_file" | awk '{print $1}')

    if [ -z "$expected_checksum" ]; then
        print_warning "Could not parse checksum from file"
        rm -f "$checksum_file"
        return 1
    fi

    # Calculate actual checksum
    if ! command_exists sha256sum; then
        print_warning "sha256sum not found, skipping SHA256 verification"
        rm -f "$checksum_file"
        return 1
    fi
    actual_checksum=$(sha256sum "$file" | awk '{print $1}')

    # Compare checksums
    if [ "$expected_checksum" = "$actual_checksum" ]; then
        print_success "SHA256 checksum verified: $actual_checksum"
        rm -f "$checksum_file"
        return 0
    else
        print_error "SHA256 checksum mismatch!"
        print_error "Expected: $expected_checksum"
        print_error "Got:      $actual_checksum"
        print_error "File may be corrupted or tampered with!"
        rm -f "$checksum_file"
        return 2
    fi
}

# Download file and verify checksums
download_file() {
    url="$1"
    output="$2"
    version="$3"
    arch="$4"

    print_info "Downloading from: $url"

    # Download the file
    if command_exists wget; then
        wget_opts="-q"
        wget $wget_opts -O "$output" "$url" || {
            print_error "Download failed"
            return 1
        }
    elif command_exists curl; then
        curl -L -# -o "$output" "$url" || {
            print_error "Download failed"
            return 1
        }
    fi

    # Construct checksum URL
    file_name="${BINARY_NAME}-linux-${arch}.tar.gz"
    sha256_url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${version}/${file_name}.sha256"

    # Try to verify SHA256 checksum
    if verify_checksum "$output" "$sha256_url"; then
        return 0
    elif [ $? -eq 2 ]; then
        # Checksum mismatch (not just missing)
        print_error "Download verification failed!"
        return 1
    else
        # Checksum file not found
        print_warning "No checksum file found in release - unable to verify download integrity"
        print_warning "Please verify manually if this is a security concern"

        # Still calculate and display local checksum for manual verification
        if command_exists sha256sum; then
            local_hash=$(sha256sum "$output" | awk '{print $1}')
            print_info "Local SHA256: $local_hash"
        fi
    fi

    return 0
}
