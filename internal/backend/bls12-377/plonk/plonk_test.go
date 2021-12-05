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

package plonk_test

import (
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"

	curve "github.com/consensys/gnark-crypto/ecc/bls12-377"

	"github.com/consensys/gnark/internal/backend/bls12-377/cs"

	bls12_377witness "github.com/consensys/gnark/internal/backend/bls12-377/witness"

	bls12_377plonk "github.com/consensys/gnark/internal/backend/bls12-377/plonk"

	"bytes"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr/kzg"
	"math/big"
	"testing"

	"github.com/consensys/gnark-crypto/ecc"
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

func referenceCircuit() (compiled.CompiledConstraintSystem, frontend.Circuit, *kzg.SRS) {
	const nbConstraints = 40000
	circuit := refCircuit{
		nbConstraints: nbConstraints,
	}
	ccs, err := frontend.Compile(curve.ID, backend.PLONK, &circuit)
	if err != nil {
		panic(err)
	}

	var good refCircuit
	good.X = (2)

	// compute expected Y
	var expectedY fr.Element
	expectedY.SetUint64(2)

	for i := 0; i < nbConstraints; i++ {
		expectedY.Mul(&expectedY, &expectedY)
	}

	good.Y = (expectedY)
	srs, err := kzg.NewSRS(ecc.NextPowerOfTwo(nbConstraints)+3, new(big.Int).SetUint64(42))
	if err != nil {
		panic(err)
	}

	return ccs, &good, srs
}

func BenchmarkSetup(b *testing.B) {
	ccs, _, srs := referenceCircuit()

	b.ResetTimer()

	b.Run("setup", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _, _ = bls12_377plonk.Setup(ccs.(*cs.SparseR1CS), srs)
		}
	})
}

func BenchmarkProver(b *testing.B) {
	ccs, _solution, srs := referenceCircuit()
	fullWitness := bls12_377witness.Witness{}
	err := fullWitness.FromFullAssignment(_solution)
	if err != nil {
		b.Fatal(err)
	}

	pk, _, err := bls12_377plonk.Setup(ccs.(*cs.SparseR1CS), srs)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = bls12_377plonk.Prove(ccs.(*cs.SparseR1CS), pk, fullWitness, backend.ProverOption{})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVerifier(b *testing.B) {
	ccs, _solution, srs := referenceCircuit()
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

	pk, vk, err := bls12_377plonk.Setup(ccs.(*cs.SparseR1CS), srs)
	if err != nil {
		b.Fatal(err)
	}

	proof, err := bls12_377plonk.Prove(ccs.(*cs.SparseR1CS), pk, fullWitness, backend.ProverOption{})
	if err != nil {
		panic(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = bls12_377plonk.Verify(proof, vk, publicWitness)
	}
}

func BenchmarkSerialization(b *testing.B) {
	ccs, _solution, srs := referenceCircuit()
	fullWitness := bls12_377witness.Witness{}
	err := fullWitness.FromFullAssignment(_solution)
	if err != nil {
		b.Fatal(err)
	}

	pk, _, err := bls12_377plonk.Setup(ccs.(*cs.SparseR1CS), srs)
	if err != nil {
		b.Fatal(err)
	}

	proof, err := bls12_377plonk.Prove(ccs.(*cs.SparseR1CS), pk, fullWitness, backend.ProverOption{})
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()

	// ---------------------------------------------------------------------------------------------
	// bls12_377plonk.ProvingKey binary serialization
	b.Run("pk: binary serialization (bls12_377plonk.ProvingKey)", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			_, _ = pk.WriteTo(&buf)
		}
	})
	b.Run("pk: binary deserialization (bls12_377plonk.ProvingKey)", func(b *testing.B) {
		var buf bytes.Buffer
		_, _ = pk.WriteTo(&buf)
		var pkReconstructed bls12_377plonk.ProvingKey
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			buf := bytes.NewBuffer(buf.Bytes())
			_, _ = pkReconstructed.ReadFrom(buf)
		}
	})
	{
		var buf bytes.Buffer
		_, _ = pk.WriteTo(&buf)
	}

	// ---------------------------------------------------------------------------------------------
	// bls12_377plonk.Proof binary serialization
	b.Run("proof: binary serialization (bls12_377plonk.Proof)", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buf bytes.Buffer
			_, _ = proof.WriteTo(&buf)
		}
	})
	b.Run("proof: binary deserialization (bls12_377plonk.Proof)", func(b *testing.B) {
		var buf bytes.Buffer
		_, _ = proof.WriteTo(&buf)
		var proofReconstructed bls12_377plonk.Proof
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

}
