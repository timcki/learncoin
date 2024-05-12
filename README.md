# learncoin

---

A cryptocurrency developed for my engineering thesis featuring a simple blockchain and confidential transactions

## Running the project

To run the cryptographic simulation run (in the root dir)

```bash
docker compose -f compose-chain-sim.yml up --build
```

To run the peer-to-peer network run the command

```bash
docker compose up --build
```

## Short technical description

This project is a basic implementation of a blockchain based cryptocurrency featuring confidentinal transactions using zero knowledge proofs.

The end-goal is to be able to generate an identity (a private EC key) and have the possibility to issue transactions to another user while concealing the amount being sent from other users. A good test of the project would be to create a network of nodes, a percentage of which would be malicious, and test if it would still be functional (producing valid blocks)

The inspiration for this thesis are mainly:

- Bitcoin - the first cryptocurrency
- Monero - the leading privacy coin
