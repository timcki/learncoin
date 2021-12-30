## Abstract
Cryptocurrencies are a new technological concept that has become increasingly popular and mainstream during the last 10-15 years. The topic of a digital peer-to-peer currency had been discussed extensively in the crypto-punk community for a long time, but the tipping point was the white-paper "Bitcoin: A Peer-to-Peer Electronic Cash System" by Satoshi Nakamoto published in 2009. Since that moment there's been a proliferation of different implementations building on the ideas presented in that document, causing advancements in various comp-sci fields, mainly cryptography, database technology and distributed computing systems.

The idea behind this thesis was to expand my practical knowledge in the field of the advanced cryptography implemented in coins that feature greater than normal transactional confidentiality such as Monero or Zcash. It ended up consisting of two separate parts: a basic p2p network and a cryptographic module allowing for fully confidential transactions with blinded receiver and sender using one-time addresses and ring signatures with key images.

---

## Brief history of cryptocurrencies
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

* ***A peer-to-peer networking layer, involving a consensus algorithm that's used to agree on a shared chain state.***
* ***A cryptographic layer allowing users to perform transactions, and exchange value.***

Even though they are distinct one cannot work without the other. This project ended up implementing both of those parts without joining them. The scope of such an endeavour would greatly exceed the workload of a typical engineering thesis and as such, has been left (as an exercise for the reader hehe) for potential future development. 
<!--- TODO: Remove the exercise for the reader bit --->
### Cryptography
#### Outline of goals
My goal with this component was to create a cryptographic scheme allowing for anonymous and untraceable exchange of value between addresses on the blockchain. The design was based on the paper *CryptoNote v2.0* by Nicolas van Saberhagen. The author of the paper set out to improve on the Bitcoin white-paper in various aspects, one of them being the cryptographic aspect of the cryptocurrency. This can be roughly summarised into 2 big improvements:
##### One-time confidential addresses
One major flow of Bitcoin is the traceability of transactions. Without much care it's possible to receive two [UTXOs](#### UTXO) into one address, in effect linking them forever as going to one recipient. With careful statistical analysis of the transaction graph one could provably uncover someone's entire transactional history, which could be disastrous for private individuals and entities.
ref: https://css.csail.mit.edu/6.858/2013/projects/jeffchan-exue-tanyaliu.pdf
ref: https://arxiv.org/pdf/2002.06403.pdf 
ref: https://arxiv.org/pdf/1502.01657.pdf
The author of the paper proposed a scheme which allows for non-interactive generation of one-time addresses. The scheme requires every user to generate 
* a private user key $(a,b)$
* a corresponding public user key $(A, B)$
subsequently publishing the public user key to allow for incoming transactions. Address generation proceeds as follows
* Sender [unpacks ](#### Public key encoding) the provided public keypair $(A, B)$
```
lrn13FdrxyjAshFzkeZscWEEGnKxeX49bdgkbWKvv5gxHRo2gFg3YyLmEBtTTA7zunGiUBpXFkcJX6Sc7SvBkKbMmxjVeF7yTFXTNQe -> (98983851ace6e4f9be668784ed5ad9b799bb98df5ac50299ffb73a6c439428ba, 4976e98410e5bec273c75738aa15e9f1512aa492e03c2b5d39a15313a8c30fd8)
```
* Sender generates `r` a random *ed25519* scalar value and computes $P$  the one-time destination address. `sha256` hashes are valid *ed25519* scalars so we can perform regular curve operations (additions, multiplication, subtraction) with them
$$ P = sha256(rA)G+B$$
* Sender compute $R$ an *ed25519* point given by the equation $R = rG$  where $G$ is the generator point of *Ed25519* elliptic curve
* Sender sends the transaction to the one time address defined as `{P,R}`. 
* Receiver checks every incoming transaction with his private user key $(a,b)$ and computes $$ P' = sha256(aR)G+B $$
* If $P=P'$  it means this transaction is meant for the receiver. This works because $$ rA = raG = arG = aR $$
* The receiver is the only one who can recover $x$  the one-time private key corresponding to our one-time address $P$ $$ x = sha256(aR)+b $$ which means he can spend this output by using $x$ in a ring-signature verifying his ownership of this address
##### Ring signatures

#### Block
The most high level building block of cryptographic structures in my project is a `Block`. Blocks serve as data structures that bunch together bundles of transactions, which allows for setting "checkpoints" for chainstate. The structure definition is as follows:
```go
// Header is the header of a block
type Header struct {
	nonce        [8]byte
	version      uint8
	blockHash    crypto.Hash
	previousHash crypto.Hash
	merkleRoot   crypto.Hash
	time         time.Time
}

type Block struct {
	header       Header
	transactions crypto.MerkleTree
}
```
A block has two distinct components, a [`Header`](### Header) and transactions stored in a [`Merkle Tree`](### Merkle Tree). The field allowing for data continuity is `previousHash` which is the `SHA-256` of the previous block of transactions. As the process of finding a block hash is very computationally intensive the deeper our block is inside the chain the harder it is to change it's contents and switch the agreed upon consenus state.
```
|------------------|          |------------------|
|      ....        |          |      ....        |
|------------------|          |------------------|
| blockHash.       | -- |     | blockHash.       |
|------------------|    |     |------------------|
| previousHash.    |    | ----| previousHash.    |
|------------------|          |------------------|
```
#### Header
Header is a component of [`Block`](#### Block). It serves as a container for metadata (protocol version, time that block was created etc.). The Merkle Tree root is one of the components which allows computing the `SHA-256` only for the header not the entire block (based on the cryptographic guarantees Merkle Trees give)
#### Merkle Tree
A Merkle Tree is a hash based tree-like data structure. It's a generalisation of the hash list. Each leaf is a piece of data, and non-leafs are hashes on their children. This gives strong integrity and verifiability guarantees ![[merkle_tree.png]] ref: https://en.wikipedia.org/wiki/Merkle_tree#/media/File:Hash_Tree.svg
Here's a table with computational complexity for Merkle Trees
| df     | Average       | Worst         |
| ------ | ------------- | ------------- |
| Space  | $O(n)$        | $O(n)$        |
| Search | $O(log_2(n))$ | $O(log_k(n))$ | 
ref: https://brilliant.org/wiki/merkle-tree/
```go
type Node struct {
	leaf    bool
	Hash    Hash
	Content *Hashable
}

type MerkleTree struct {
	nodes []Node
	depth uint
}
```
Here is the definition of a Merkle Tree in my thesis. As it's essentially a binary tree I decided to opt for an array representation as it's more concise and less allocation intensive when compared to a typical linked list-like approach
```go
// NewMerkleTree constructs a binary tree in which
// the leaves are transactions and parents are subsequent
// hashes. Since we know the depth at which the children will start
func NewMerkleTree(elements []Hashable) (*MerkleTree, error) {
	var pow float64
	elements, pow = fillElements(elements)
	// Tree size is sum(2^n for n from pow to 1)
	// so also 2^(n+1) - 1
	size := int(math.Pow(2, pow+1) - 1)
	tree := make([]Node, size)

	// Calculate the base index for level of leaves
	baseIndex := int(math.Pow(2, pow)) - 1
	// Fill all of the leaves with txs
	for i, element := range elements {
		hash, err := element.Hash()
		if err != nil {
			return nil, err
		}
		tree[baseIndex+i] = Node{leaf: true, Hash: hash, Content: &element}
	}
	// Fill all parent nodes until root
	for i := baseIndex - 1; i >= 0; i-- {
		childrenHash, err := HashData(append(tree[2*i+1].Hash, tree[2*i+2].Hash...))
		if err != nil {
			return nil, err
		}
		tree[i] = Node{leaf: false, Hash: childrenHash, Content: nil}
	}
	return &MerkleTree{tree, uint(baseIndex)}, nil
}
```
Since we can't guarantee that we'll have exactly $2^n$ elements we also need a function to fill out the $2^n - elements$  leafs left.
```go
// fillElements rounds the number of leafs to the nearest
// power of 2 greater than the length
func fillElements(el []Hashable) ([]Hashable, float64) {
	l := len(el)
	// Calculate the nearest power of two greater than the len of our tx list
	pow := math.Ceil(math.Log2(float64(l)))
	for i := 0; i < int(math.Pow(2, pow))-l; i++ {
		el = append(el, EmptyLeaf(0))
	}
	return el, pow
}
```
When we're ready to create a block we can retrieve the top hash and embed it into our block header
```go
func (m MerkleTree) RootHash() Hash {
	return m.nodes[0].Hash
}
```

#### Transactions
Transactions are the most important component of any cryptocurrency as they are the *raison d'etre*  for them. A transaction in my 
#### Public key encoding
