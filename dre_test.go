package otr4

import (
	"crypto/rand"

	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type OTR4Suite struct{}

var _ = Suite(&OTR4Suite{})

func (s *OTR4Suite) Test_Auth(c *C) {
	message := []byte("our message")
	out, err := auth(fixedRand(randAuthData), testPubA, testPubB, testPubC, testSec, message)

	c.Assert(out, DeepEquals, testSigma)
	c.Assert(err, IsNil)

	r := make([]byte, 270)
	out, err = auth(fixedRand(r), testPubA, testPubB, testPubC, testSec, message)

	c.Assert(err, ErrorMatches, ".*cannot source enough entropy")
	c.Assert(out, IsNil)

	r = make([]byte, 56)
	out, err = auth(fixedRand(r), testPubA, testPubB, testPubC, testSec, message)

	c.Assert(err, ErrorMatches, ".*cannot source enough entropy")
	c.Assert(out, IsNil)
}

func (s *OTR4Suite) Test_Verify(c *C) {
	message := []byte("our message")

	b := verify(testPubA, testPubB, testPubC, testSigma, message)

	c.Assert(b, Equals, true)
}

func (s *OTR4Suite) Test_VerifyAndAuth(c *C) {
	message := []byte("hello, I am a message")
	sigma, _ := auth(rand.Reader, testPubA, testPubB, testPubC, testSec, message)
	ver := verify(testPubA, testPubB, testPubC, sigma, message)
	c.Assert(ver, Equals, true)

	fakeMessage := []byte("fake message")
	ver = verify(testPubA, testPubB, testPubC, sigma, fakeMessage)
	c.Assert(ver, Equals, false)

	ver = verify(testPubB, testPubB, testPubC, sigma, message)
	c.Assert(ver, Equals, false)

	ver = verify(testPubA, testPubA, testPubC, sigma, message)
	c.Assert(ver, Equals, false)

	ver = verify(testPubA, testPubB, testPubB, sigma, message)
	c.Assert(ver, Equals, false)

	ver = verify(testPubA, testPubB, testPubC, testSigma, message)
	c.Assert(ver, Equals, false)
}

func (s *OTR4Suite) Test_DREnc(c *C) {
	drMessage := new(drMessage)
	err := drMessage.drEnc(keyMessage, fixedRand(randDREData), pubA, pubB)

	c.Assert(drMessage.cipher, DeepEquals, expDRMessage.cipher)
	c.Assert(drMessage.proof, DeepEquals, expDRMessage.proof)
	c.Assert(err, IsNil)
}

func (s *OTR4Suite) Test_DRDec(c *C) {
	m, err := expDRMessage.drDec(pubA, pubB, privA, 1)

	c.Assert(m, DeepEquals, keyMessage)
	c.Assert(err, IsNil)
}

func (s *OTR4Suite) Test_DREncryptAndDecrypt(c *C) {
	message := []byte{
		0xfd, 0xf1, 0x18, 0xbf, 0x8e, 0xc9, 0x64, 0xc7,
		0x94, 0x46, 0x49, 0xda, 0xcd, 0xac, 0x2c, 0xff,
		0x72, 0x5e, 0xb7, 0x61, 0x46, 0xf1, 0x93, 0xa6,
		0x70, 0x81, 0x64, 0x37, 0x7c, 0xec, 0x6c, 0xe5,
		0xc6, 0x8d, 0x8f, 0xa0, 0x43, 0x23, 0x45, 0x33,
		0x73, 0x79, 0xa6, 0x48, 0x57, 0xbb, 0x0f, 0x70,
		0x63, 0x8c, 0x62, 0x26, 0x9e, 0x17, 0x5d, 0x22,
	}

	priv1, pub1, err := deriveCramerShoupKeys(rand.Reader)
	priv2, pub2, err := deriveCramerShoupKeys(rand.Reader)

	drMessage := &drMessage{}
	err = drMessage.drEnc(message, rand.Reader, pub1, pub2)

	expMessage1, err := drMessage.drDec(pub1, pub2, priv1, 1)
	expMessage2, err := drMessage.drDec(pub1, pub2, priv2, 2)
	c.Assert(expMessage1, DeepEquals, message)
	c.Assert(expMessage2, DeepEquals, message)
	c.Assert(err, IsNil)
}
