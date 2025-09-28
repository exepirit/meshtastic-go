package meshtastic

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/binary"
	"github.com/exepirit/meshtastic-go/pkg/meshtastic/proto"
	protobuf "google.golang.org/protobuf/proto"
)

// DecryptPSK decrypts the encrypted payload of a MeshPacket using the provided AES cipher block.
func DecryptPSK(packet *proto.MeshPacket, block cipher.Block) (*proto.Data, error) {
	encrypted := packet.GetEncrypted()
	decrypted := make([]byte, len(encrypted))

	nonce := make([]byte, 16)
	binary.LittleEndian.PutUint32(nonce[0:], packet.GetId())
	binary.LittleEndian.PutUint32(nonce[8:], packet.GetFrom())
	cipher.NewCTR(block, nonce).XORKeyStream(decrypted, encrypted)

	decryptedData := new(proto.Data)
	if err := protobuf.Unmarshal(decrypted, decryptedData); err != nil {
		return nil, ErrInvalidPacketFormat
	}
	return decryptedData, nil
}

// DecodeCipherKeyBase64 converts a base64-encoded string into an AES cipher block.
func DecodeCipherKeyBase64(key string) (cipher.Block, error) {
	bytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}

	c, err := aes.NewCipher(bytes)
	if err != nil {
		return nil, err
	}
	return c, nil
}
