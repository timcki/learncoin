## Abstract
Cryptocurrencies are a new technological concept that has become increasingly popular and mainstream during the last 10-15 years. The topic of a digital peer-to-peer currency had been discussed extensively in the crypto-punk community for a long time, but the tipping point was the white-paper "Bitcoin: A Peer-to-Peer Electronic Cash System" by Satoshi Nakamoto published in 2009. Since that moment there's been a proliferation of different implementations building on the ideas presented in that document, causing advancements in various comp-sci fields, mainly cryptography, database technology and distributed computing systems.

The idea behind this thesis was to expand my practical knowledge in the field of the advanced cryptography implemented in coins that feature greater than normal transactional confidentiality such as Monero or Zcash. It ended up consisting of two separate parts: a basic p2p network and a cryptographic module allowing for fully confidential transactions with blinded receiver and sender using one-time addresses and ring signatures with key images.


## Brief history of cryptcurrencies
In this section I'd like to introduce a short history of concepts that came before Bitcoin which led to it's creation, as well as showing some interesting developments that came afterwards
### Digicash
The first company implementing a "digital cash" system was probably Digicash founded in the late 80 by David Chaum. It promised to implement anonymous online payments, but failed because of the centralization of it's components. Nonetheless it introduced the concept of \textbf{blind signatures} which later got to be known as addresses in more recent cryptocurrencies.
### Bit-gold
Big-gold was an idea introduced by cryptographer Nick Szabo in 1998. It was the first real, decentralized, cryptographicaly sound digital cash concept which never got implemented. It introduced the concept of Proof-of-Work which was arguably it's biggest breakthrough. The project was so similar to Bitcoin that many suspected Satoshi Nakamoto to be Nick Szabo. He later denied those allegations.
### Hashcash
Hashcash is a proof-of-work algorithm invented by Adam Back in order to combat spam and DDoS attacks. It was used in the Bitcoin white-paper as the default algorithm for mining because of it's simplicity and elegance. Adam Back has been heavily involved in the development of Bitcoin since the very beginning through his company: Blockstream.
### Ethereum
Vitalik Buterin created Ethereum in 2015 after years of work with his close team. Ethereum is a cryptocurrency that implements a general purpose, Turing complete virtual machine inside it's binary (the EVM). This allows much more complicated computation then Bitcoin, which introduced the concept of smart contracts: programs stored on the blockchain that execute code when given conditions are met. Use-cases for this involve decentralised currency-exchanges without third parties and non-fungible tokens.
### Monero
The origin of Monero (Esperanto for Coin) was a white-paper from 2013 called "CryptoNote v2" authored by Nicolas van Saberhagen - a fictional character. The original paper expanded on the ideas introduced by Satoshi Nakamoto by introducing confidential transactions. Monero built on those concepts with further research from academia, and is now the leading project in terms of privacy preserving e-cash systems along with Z-Cash.

## Implementation of my cryptocurrency
The goal of the project was to understand the inner workings of modern cryptocurrencies. Those can be divided into 2 distinct layers:
* A peer-to-peer networking layer, involving a consensus algorithm that's used to agree on a shared chain state.
  * A cryptographic layer allowing users to perform transactions, and exchange value.
Even though they are distinct one cannot work without the other. This project ended up implementing both of those parts without joining them. The scope of such an endeavour would greatly exceed the workload of a typical engineering thesis and as such, has been left for potential future development. 
### Cryptography
The most high level building block of the cryptography in my project is a `Transaction Block`. The structure definition is as follows:
```go
// Header is the header of a block
type Header struct {
	version      uint8
	previousHash crypto.Hash
	merkleRoot   crypto.Hash
	time         time.Time
}

type Block struct {
	header       Header
	transactions crypto.MerkleTree
}
```
As the name entails it serves as a container for a group of transactions. Those are being held in a Merkle Tree
### Merkle Tree
