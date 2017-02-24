package otr4

import (
	"io"

	"golang.org/x/crypto/sha3"

	"github.com/twstrike/ed448"
)

// XXX: use bytes?
type cramerShoupPrivateKey struct {
	x1, x2, y1, y2, z ed448.Scalar
}

type cramerShoupPublicKey struct {
	c, d, h ed448.Point
}

type cramerShoupMessage struct {
	u1, u2, e, v ed448.Point
}

func deriveCramerShoupPrivKey(rand io.Reader) (*cramerShoupPrivateKey, error) {
	priv := &cramerShoupPrivateKey{}
	var err error
	priv.x1, err = randLongTermScalar(rand)
	if err != nil {
		return priv, err
	}
	priv.x2, err = randLongTermScalar(rand)
	if err != nil {
		return priv, err
	}
	priv.y1, err = randLongTermScalar(rand)
	if err != nil {
		return priv, err
	}
	priv.y2, err = randLongTermScalar(rand)
	if err != nil {
		return priv, err
	}
	priv.z, err = randLongTermScalar(rand)
	if err != nil {
		return priv, err
	}
	return priv, nil
}

// TODO: HANDLE ERROR
func deriveCramerShoupKeys(rand io.Reader) (*cramerShoupPrivateKey, *cramerShoupPublicKey, error) {
	priv, _ := deriveCramerShoupPrivKey(rand)
	pub := &cramerShoupPublicKey{}
	pub.c = ed448.DoubleScalarMul(ed448.BasePoint, g2, priv.x1, priv.x2)
	pub.d = ed448.DoubleScalarMul(ed448.BasePoint, g2, priv.y1, priv.y2)
	pub.h = ed448.PointScalarMul(ed448.BasePoint, priv.z)
	return priv, pub, nil
}

func (csm *cramerShoupMessage) cramerShoupEnc(message []byte, rand io.Reader, pub *cramerShoupPublicKey) error {
	r, err := randScalar(rand)
	if err != nil {
		return err
	}

	// u = G1*r, u2 = G2*r
	csm.u1 = ed448.PointScalarMul(ed448.BasePoint, r)
	csm.u2 = ed448.PointScalarMul(g2, r)

	// e = (h*r) + m
	m := ed448.NewPointFromBytes(nil)
	m.Decode(message, false)
	csm.e = ed448.PointScalarMul(pub.h, r)
	csm.e.Add(csm.e, m)

	// α = H(u1,u2,e)
	al := concat(csm.u1, csm.u2, csm.e)
	hash := sha3.NewShake256()
	hash.Write(al)
	var alpha [fieldBytes]byte
	hash.Read(alpha[:])

	// a = c * r
	// b = d*(r * alpha)
	// v = s + t
	a := ed448.PointScalarMul(pub.c, r)
	b := ed448.PointScalarMul(pub.d, r)
	b = ed448.PointScalarMul(b, ed448.NewDecafScalar(alpha[:]))
	csm.v = ed448.NewPointFromBytes(nil)
	csm.v.Add(a, b)

	return nil
}

func (csm *cramerShoupMessage) cramerShoupDec(priv *cramerShoupPrivateKey) (message []byte, err error) {

	// alpha = H(u1,u2,e)
	al := concat(csm.u1, csm.u2, csm.e)
	hash := sha3.NewShake256()
	hash.Write(al)
	var alpha [56]byte
	hash.Read(alpha[:])

	// (u1*(x1+y1*alpha) +u2*(x2+ y2*alpha) == v
	// a = (u1*x1)+(u2*x2)
	a := ed448.DoubleScalarMul(csm.u1, csm.u2, priv.x1, priv.x2)
	// b = (u1*y1)+(u2*y2)
	b := ed448.DoubleScalarMul(csm.u1, csm.u2, priv.y1, priv.y2)
	v0 := ed448.PointScalarMul(b, ed448.NewDecafScalar(alpha[:]))
	v0.Add(a, v0)

	valid := v0.Equals(csm.v)

	if !valid {
		return nil, newOtrError("verification of cipher failed")
	}

	// m = e - u1*z
	m := ed448.PointScalarMul(csm.u1, priv.z)
	m.Sub(csm.e, m)
	message = m.Encode()
	return
}
