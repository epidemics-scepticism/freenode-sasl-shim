# freenode-sasl-shim
Provides a simple shim that allows clients with no SASL support to access Freenode's onion.

Listens on 127.0.0.1:17649 and reads a TLS keypair from `sasl.crt` and `sasl.key`.

Connects to Freenode's .onion over Tor using the crt/key as a client certificate.

Negotiates EXTERNAL SASL with Freenode using the crt/key.

First make a keypair:

    openssl req -x509 -new -newkey rsa:4096 -sha256 -days 1000 -nodes -out sasl.crt -keyout sasl.key
    # This creates the SASL certificate and key
    openssl x509 -in sasl.crt -outform der | sha1sum -b | cut -d' ' -f1
    # This will print out it's fingerprint

Then login to Freenode *without* the .onion and identify with nickserv, then add your fingerprint:

    /msg nickserv CERT ADD 3683af4ba1238aaaea25a4126d218e0377a59c00
    # Assuming your fingerprint you received was "3683af4ba1238aaaea25a4126d218e0377a59c00"

Now launch `freenode-sasl-shim`, where `sasl.crt` and `sasl.key` should be located in $PWD.

Configure your client to use an IRC server at 127.0.0.1:17649 (choose no proxy for this account), and connect.
