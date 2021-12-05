// Copyright 2020 ConsenSys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by gnark DO NOT EDIT

package groth16_test

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"

	curve "github.com/consensys/gnark-crypto/ecc/bls12-377"

	"github.com/consensys/gnark/internal/backend/bls12-377/cs"

	bls12_377witness "github.com/consensys/gnark/internal/backend/bls12-377/witness"

	"bytes"
	bls12_377groth16 "github.com/consensys/gnark/internal/backend/bls12-377/groth16"
	"testing"

	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/frontend"
	frontendcs "github.com/consensys/gnark/frontend/cs"
	"github.com/consensys/gnark/internal/backend/compiled"
)

//--------------------//
//     benches		  //
//--------------------//

type refCircuit struct {
	nbConstraints int
	X             frontendcs.Variable
	Y             frontendcs.Variable `gnark:",public"`
}

func (circuit *refCircuit) Define(api frontend.API) error {
	for i := 0; i < circuit.nbConstraints; i++ {
		circuit.X = api.Mul(circuit.X, circuit.X)
	}
	api.AssertIsEqual(circuit.X, circuit.Y)
	return nil
}

func referenceCircuit() (compiled.CompiledConstraintSystem, frontend.Circuit) {
	const nbConstraints = 40000
	circuit := refCircuit{
		nbConstraints: nbConstraints,
	}
	r1cs, err := frontend.Compile(curve.ID, backend.GROTH16, &circuit)
	if err != nil {
		panic(err)
	}

	var good refCircuit
	good.X = 2

	// compute expected Y
	var expectedY fr.Element
	expectedY.SetUint64(2)

	for i := 0; i < nbConstraints; i++ {
		expectedY.Mul(&expectedY, &expectedY)
	}

	good.Y = (expectedY)

	return r1cs, &good
}

func BenchmarkSetup(b *testing.B) {
	r1cs, _ := referenceCircuit()

	var pk bls12_377groth16.ProvingKey
	var vk bls12_377groth16.VerifyingKey
	b.ResetTimer()

	b.Run("setup", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			bls12_377groth16.Setup(r1cs.(*cs.R1CS), &pk, &vk)
		}
	})
}

func BenchmarkProver(b *testing.B) {
	r1cs, _solution := referenceCircuit()
	fullWitness := bls12_377witness.Witness{}
	err := fullWitness.FromFullAssignment(_solution)
	if err != nil {
		b.Fatal(err)
	}

	var pk bls12_377groth16.ProvingKey
	bls12_377groth16.DummySetup(r1cs.(*cs.R1CS), &pk)

	b.ResetTimer()
	b.Run("prover", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = bls12_377groth16.Prove(r1cs.(*cs.R1CS), &pk, fullWitness, backend.ProverOption{})
		}
	})
}

func BenchmarkVerifier(b *testing.B) {
	r1cs, _solution := referenceCircuit()
	fullWitness := bls12_377witness.Witness{}
	err := fullWitness.FromFullAssignment(_solution)
	if err != nil {
		b.Fatal(err)
	}
	publicWitness := bls12_377witness.Witness{}
	err = publicWitness.FromPublicAssignment(_solution)
	if err != nil {
		b.Fatal(err)
	}

	var pk bls12_377groth16.ProvingKey
	var vk bls12_377groth16.VerifyingKey
	bls12_377groth16.Setup(r1cs.(*cs.R1CS), &pk, &vk)
	proof, err := bls12_377groth16.Prove(r1cs.(*cs.R1CS), &pk, fullWitness, backend.ProverOption{})
	if err != nil {
		panic(err)
	}

	b.ResetTimer()
	b.Run("verifier", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = bls12_377groth16.Verify(proof, &vk, publicWitness)
		}
	})
}

func BenchmarkProofSerialization(b *testing.B) {
	r1cs, _solution := referenceCircuit()
	fullWitness := bls12_377witness.Witness{}
	err := fullWitness.FromFullAssignment(_solution)
	if err != nil {
		b.Fatal(err)
	}

	var pk bls12_377groth16.ProvingKey
	var vk bls12_377groth16.VerifyingKey
	bls12_377groth16.Setup(r1cs.(*cs.R1CS), &pk, &vk)
	proof, err := bls12_377groth16.Prove(r1cs.(*cs.R1CS), &pk, fullWitness, backend.ProverOption{})
	if err != nil {
		panic(err)
	}

	b.ReportAllocs()

	// ---------------------------------------------------------------------------------------------
	// bls12_377groth16.Proof binary serialization
	b.Run("proof: binary serialization (bls12_377groth16.Proof)", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			_, _ = proof.WriteTo(&buf)
		}
	})
	b.Run("proof: binary deserialization (bls12_377groth16.Proof)", func(b *testing.B) {
		var buf bytes.Buffer
		_, _ = proof.WriteTo(&buf)
		var proofReconstructed bls12_377groth16.Proof
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(buf.Bytes())
			_, _ = proofReconstructed.ReadFrom(buf)
		}
	})
	{
		var buf bytes.Buffer
		_, _ = proof.WriteTo(&buf)
	}

	// ---------------------------------------------------------------------------------------------
	// bls12_377groth16.Proof binary serialization (uncompressed)
	b.Run("proof: binary raw serialization (bls12_377groth16.Proof)", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			_, _ = proof.WriteRawTo(&buf)
		}
	})
	b.Run("proof: binary raw deserialization (bls12_377groth16.Proof)", func(b *testing.B) {
		var buf bytes.Buffer
		_, _ = proof.WriteRawTo(&buf)
		var proofReconstructed bls12_377groth16.Proof
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(buf.Bytes())
			_, _ = proofReconstructed.ReadFrom(buf)
		}
	})
	{
		var buf bytes.Buffer
		_, _ = proof.WriteRawTo(&buf)
	}

}

func BenchmarkProvingKeySerialization(b *testing.B) {
	r1cs, _ := referenceCircuit()

	var pk bls12_377groth16.ProvingKey
	bls12_377groth16.DummySetup(r1cs.(*cs.R1CS), &pk)

	var buf bytes.Buffer
	// grow the buffer once
	pk.WriteTo(&buf)

	b.ResetTimer()
	b.Run("pk_serialize_compressed", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf.Reset()
			pk.WriteTo(&buf)
		}
	})

	compressedBytes := buf.Bytes()
	b.ResetTimer()
	b.Run("pk_deserialize_compressed_safe", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			pk.ReadFrom(bytes.NewReader(compressedBytes))
		}
	})

	b.ResetTimer()
	b.Run("pk_deserialize_compressed_unsafe", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			pk.UnsafeReadFrom(bytes.NewReader(compressedBytes))
		}
	})

	b.ResetTimer()
	b.Run("pk_serialize_raw", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf.Reset()
			pk.WriteRawTo(&buf)
		}
	})

	rawBytes := buf.Bytes()
	b.ResetTimer()
	b.Run("pk_deserialize_raw_safe", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			pk.ReadFrom(bytes.NewReader(rawBytes))
		}
	})

	b.ResetTimer()
	b.Run("pk_deserialize_raw_unsafe", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			pk.UnsafeReadFrom(bytes.NewReader(rawBytes))
		}
	})
}
