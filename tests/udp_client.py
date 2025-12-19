import socket

def build_tftp_rrq(filename: str, mode: str = "octet") -> bytes:
    if "\x00" in filename or "\x00" in mode:
        raise ValueError("Filename and mode must not contain null bytes")

    opcode = b"\x00\x01"  # RRQ
    return (
        opcode +
        filename.encode("ascii") + b"\x00" +
        mode.encode("ascii") + b"\x00"
    )

def build_tftp_ack(blocknum: bytes) -> bytes:
    return b"\x00\x04" + blocknum

def send_and_receive_udp(
    host: str,
    port: int,
    data: bytes,
    timeout: float = 2.0,
    bufsize: int = 4096,
) -> bytes | None:
    with socket.socket(socket.AF_INET, socket.SOCK_DGRAM) as sock:
        sock.settimeout(timeout)

        sock.sendto(data, (host, port))
        file = b""
        try:
            response, addr = sock.recvfrom(bufsize)
            # first 4 bytes are opcode & block num, do not add to file.
            file += response[4:]
            while len(response) == 516:
                blocknum = response[2:4]
                ack = build_tftp_ack(blocknum)
                sock.sendto(ack, addr)
                response, addr = sock.recvfrom(bufsize)
                file += response[4:]
            print("final:")
            print(file)
            with open("resp.txt", 'w') as f:
                f.write(file.decode('utf-8'))
            return response
        except socket.timeout:
            print("No response (timeout)")
            return None

if __name__ == "__main__":
    target_host = "127.0.0.1"
    target_port = 8069

    rrq = build_tftp_rrq("mini.txt")

    print("Sending RRQ:", rrq)

    response = send_and_receive_udp(target_host, target_port, rrq)

    if response:
        print("done")

