# LANCopy — `lancp`

[![Go Report Card](https://goreportcard.com/badge/github.com/nchaloult/lancp)](https://goreportcard.com/report/github.com/nchaloult/lancp)

Easily transfer files securely between two machines on the same local network. Similar to `scp` and `rsync`, but more convenient to use.

![Demo](https://user-images.githubusercontent.com/31291920/114258855-8840ed80-9997-11eb-882e-962e21b0a8c3.gif)

## How It Works

`lancp` helps two machines on the same network find each other through a **device discovery handshake**, establishes a secure **TLS connection** between them, then sends a file over that connection.

`lancp` never reaches out to the open Internet, so it will work between two machines as long they are both connected to the same router.

### Device Discovery Handshake

The device discovery handshake is composed of two UDP messages between a machine that wants to receive a file and a machine that wants to send a file.

First, the receiver machine begins listening for a sender machine to reach out. It displays a passphrase on screen. Any sender machine that wants to establish a connection must reach out with this passphrase. If anyone reaches out with the wrong passphrase, the receiver machine immediately stops listening for more messages, and the `lancp` process terminates. This prevents anyone who the receiver has not shared the passphrase with from sending them a file. Notice that the receiver never notifies the sender that their passphrase was incorrect. This prevents attempts from anyone else on the local network from finding out if there is a receiver who's listening.

When a sender wants to reach out to a listening receiver, they send a [UDP broadcast message](https://en.wikipedia.org/wiki/Broadcast_address) to the router they're connected to. The payload of this message is the sender's guess at the receiver's passphrase. Because broadcast messages are a characteristic of the UDP protocol, all routers know how to send those messages to every device connected to them. `lancp` takes advantage of this to enable a sender to reach out to a receiver without knowing that receiver's local IP address.

Immediately after sending a broadcast message, the sender assumes that its guess is correct, chooses another passphrase of its own, and displays it on screen. It then waits for the receiver to respond with a guess at that passphrase. Like before, if the receiver responds with an incorrect guess, then the sender immediately stops listening for more messages, and the `lancp` process terminates.

Meanwhile, if the sender's passphrase guess from the original broadcast message was correct, the receiver will respond to the sender with a UDP message with its guess at the sender's passphrase.

At this point, both the sender and receiver have exchanged passphrases and verified each other's identities. Now they're ready to establish an encrypted connection and exchange a file.

### Preparing for a TLS Connection

After the device discovery handshake is finished, the sender and receiver machines are ready to establish an encrypted connection. TLS is a good protocol for this since its cipher suite takes care of everything for us, from exchanging a shared secret key to encrypt messages with, to authenticating messages once they're received, and lots in between.

In order to perform the TLS handshake, though, the receiver needs to present a certificate to the sender. That certificate contains a public key, and information that certifies the identity of the corresponding private key's owner. Normally, a universally-trusted certificate authority signs the certificate to verify the owner, but in our case, we've already certified the sender and receiver's identities. This, along with the fact that we're only using TLS as a vehicle for setting up an encrypted connection, means that we don't care who signs this certificate.

Because we aren't worried about who signs the receiver's certificate, the simplest option is to have the receiver sign their own certificate instead of getting an official, trusted certificate authority to do it. By doing this, the receiver produces a self-signed certificate.

So, to prepare for establishing an encrypted connection, the receiver generates a self-signed SSL certificate, and sets up a TCP listener. The sender reaches out and attempts to establish a TCP connection. The receiver then sends the sender this certificate (in plaintext, but that's fine) and closes the TCP connection.

At this point, the sender has everything that they need to reach out to the receiver once more and establish a TLS connection.

## Motivation

Plenty of tools and services exist that let you share files between multiple computers, but I struggled to find one that was a perfect fit for me. Many of them are meant for general-purpose file sharing, collaborating with others, or maintaining backups of your stuff, but I just wanted to transfer a file between my Mac laptop and my Linux desktop every once in a while. I basically wanted AirDrop, but for any computer.

Before making `lancp`, I tried

- **Google Drive** and **Dropbox**, but their desktop apps are cumbersome for my needs. You have to
    - Wait for the apps to recognize that you've added a new file to one of your synced folders
    - Wait for that file to upload to their platform
    - Wait for your other machine to download that file
    - Remember to delete it from that synced folder afterwards
- **[SyncThing](https://syncthing.net/)**, but
    - You have to wait for it to recognize that you've added a new file to one of your synced folders
        - You can force its file system watcher to kick in with the CLI, but that's not a seamless solution
    - Like Google Drive and Dropbox, it's really meant for keeping entire folders synced across machines rather than for quickly moving one file around
- **[magic-wormhole](https://github.com/magic-wormhole/magic-wormhole)**, but that depends on its [transit-relay server](https://github.com/magic-wormhole/magic-wormhole-transit-relay) (similar to a STUN server)
    - This means you need a connection to the open Internet. If your ISP fails you, you can't use `magic-wormhole`
    - Nit-picky, but: you have to trust the developers and maintainers of the transit-relay server
        - How can you be sure that the open source code is actually what's running in production?
- Tools like **[scp](https://en.wikipedia.org/wiki/Secure_copy_protocol)** and **[rsync](https://en.wikipedia.org/wiki/Rsync)**, but you need to know the local IP address of the recipient machine
    - I'm not gonna remember those lol
    - The local IP addresses for my devices on my home network change every few weeks or so
        - I can't figure out how to assign static addresses in my router's control panel :(

Working on `lancp` was also a good excuse for me to learn more about computer networking. While I didn't need to use, or work directly with, all of these things to build `lancp`, I had a blast exploring and researching topics like AES, elliptic curve cryptography, X.509 certificates, certificate authorities, secret exchanges like Diffie-Hellman, and the differences between hashed data and message authentication codes (like HMACs).
