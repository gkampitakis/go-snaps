package examples

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
)

func TestMain(t *testing.M) {
	v := t.Run()

	snaps.Clean()

	os.Exit(v)
}

func TestSimple(t *testing.T) {
	t.Run("should make an int snapshot", func(t *testing.T) {
		snaps.MatchSnapshot(t, 5)
	})

	t.Run("should make a string snapshot", func(t *testing.T) {
		snaps.MatchSnapshot(t, "string snapshot")
	})

	t.Run("should make a map snapshot", func(t *testing.T) {
		// snaps.Skip(t)
		m := map[string]interface{}{
			"mock-0": "value",
			"mock-1": 2,
			"mock-2": func() {},
			"mock-3": float32(10.4),
		}

		snaps.MatchSnapshot(t, m)
	})

	t.Run("should make multiple entries in snapshot", func(t *testing.T) {
		snaps.MatchSnapshot(t, 15, 10, 20, 25)
	})

	t.Run("should make create multiple snapshot", func(t *testing.T) {
		// snaps.Skip(t)
		snaps.MatchSnapshot(t, 1000)
		snaps.MatchSnapshot(t, "another snapshot")
		snaps.MatchSnapshot(t, `{
			"user": "gkampitakis",
			"id": 1234567AAA,
			"data": [ ]
		}`)
	})

	t.Run("nest", func(t *testing.T) {
		t.Run("more", func(t *testing.T) {
			t.Run("one more nested test", func(t *testing.T) {
				snaps.MatchSnapshot(t, "it's okay")
			})
		})
	})

	t.Run(".*", func(t *testing.T) {
		snaps.MatchSnapshot(t, "ignor dasdase regex patterns on names")
	})
}

func TestSimpleTable(t *testing.T) {
	type testCases struct {
		description string
		input       interface{}
	}

	for _, scenario := range []testCases{
		{
			description: "string",
			input:       "input",
		},
		{
			description: "integer",
			input:       10,
		},
		{
			description: "map",
			input: map[string]interface{}{
				"test": func() {},
			},
		},
		{
			description: "buffer",
			input:       bytes.NewBufferString("Buffer string"),
		},
	} {
		t.Run(scenario.description, func(t *testing.T) {
			snaps.MatchSnapshot(t, scenario.input)
		})
	}
}

func TestRefactorDiff(t *testing.T) {
	t.Run("should show the diff for newline", func(t *testing.T) {
		snaps.Skip(t)
		// BUG: this is a special case where the diff char is \n
		snaps.MatchSnapshot(t, "good morning more and more")

		snaps.MatchSnapshot(t, `

		
		
		`)
	})

	t.Run("should test single diff", func(t *testing.T) {
		snaps.Skip(t)

		snaps.MatchSnapshot(t, strings.Repeat("geore", 1000))
	})

	t.Run("long unchanged diffs", func(t *testing.T) {
		snaps.MatchSnapshot(t, `Pretium Ac Erat Nostra Id Purus Habitasse Ut Suscipit Nisl Interdum Pharetra Tortor Taciti Aliquet Luctus Proin Ultricies Nullam Tempus Quisque Mi Lacus Ut 
		Tellus Fringilla Dictum Consequat Cras Aenean Nunc Congue Vehicula Erat Velit Torquent Eget Bibendum Dictumst Curabitur Urna Ultricies Dictum Condimentum Bibendum Lectus Dictum 
		Nisi Ligula Magna Nam Ut Justo Tincidunt Hendrerit Erat Elementum Orci Donec Eu Laoreet Porta Tempus Vestibulum Porta Ut Accumsan Dictum Facilisis Massa Adipiscing Donec Sapien Eros Sapien Sagittis Potenti Venenatis A Platea Consequat Feugiat Integer 
		Curabitur Fermentum Consequat Torquent Etiam Tempor Lorem Eros Etiam Elementum Quisque Fames Nullam Ipsum Lorem Sem Eu Mauris Semper Malesuada Tellus Pulvinar Orci Gravida Amet Hendrerit Habitant Metus Blandit Ullamcorper Habitant 
		Proin Enim Phasellus Malesuada Ut Lectus Senectus Molestie Metus Faucibus Augue Arcu Class Sapien Sit Lacus Lacinia Dapibus Ligula Cras Duis Rutrum Lorem Feugiat Nec Enim Massa Rutrum Arcu Himenaeos Volutpat Hac Lacus Purus Tristique Vulputate 
		Phasellus Potenti Mattis Lobortis Id Tortor Leo Duis Enim Quis Ligula Metus Vulputate Vestibulum Euismod Venenatis 
		Sapien Vivamus Sagittis Aliquet Vivamus Vehicula Aliquam Ligula Dui Litora Bibendum Gravida Praesent Scelerisque Velit Sollicitudin Lacus Quam Nibh Ultricies Dictum Ullamcorper Facilisis Lorem Nam Per Proin Rutrum Dolor 
		Tristique Condimentum Sollicitudin Nulla Per Tortor Auctor Tortor Etiam Vulputate Quis Duis Scelerisque Sollicitudin Interdum Volutpat Cras Porta Consectetur Magna Lacus Odio Consectetur Potenti Conubia Massa Fringilla 
		Commodo Luctus Pharetra Bibendum Vitae Primis Tellus Fames Primis Dictum Duis Integer Molestie Ullamcorper Euismod At Ante Curabitur Eu Hendrerit Purus 
		Mauris Leo Vulputate Semper Nisl Felis Volutpat  Lorem Fringilla Rhoncus Fermentum Tempor Aenean Habitasse Lobortis Vitae Lorem Ultrices Auctor Litora Eu Sagittis Etiam Lobortis Pulvinar Ac Quisque Augue At Diam 
		Hendrerit Enim Tortor Leo Vulputate Adipiscing Curae Lectus Volutpat Tellus Nisl Dapibus Augue Himenaeos Arcu Semper Ornare Quisque Metus Euismod Luctus Facilisis Neque Torquent Donec Aptent Eros Fringilla Nec Leo 
		Bibendum Quis Eget Ac Ultrices Etiam Dapibus Lectus Dictumst Pretium Ipsum Nisi Nulla Arcu Diam Porta Netus Aliquam Purus In Phasellus Dictum Condimentum Mattis 
		ssa Volutpat Torquent Aenean Adipiscing Arcu Etiam Nam Dui Fermentum Ultrices Elementum Enim Quis Morbi Orci Per Donec Rutrum Porttitor Tempus Interdum Vitae Consequat A Dictum Magna Amet Viverra Proin Aptent Nullam Pellentesque
		Faucibus Vulputate Iaculis Condimentum Curae Vestibulum Consectetur Class Ullamcorper Commodo Gravida Mi Eu Massa Blandit Nisl Primis Sapien Maecenas Condimentum Dictumst Sodales Aenean Porta Tempus Ipsum Crack Iaculis Orci 
		Tempus Arcu Blandit Praesent Ut Ac Viverra Ut Diam Himenaeos 
		Lacus Mattis Pharetra Inceptos Nullam In Nam A Sollicitudin Egestas Ultrices Duis Non Donec Integer Eget Cursus Massa Curae Facilisis Sed Ipsum Donec Integer 
		Gravida Vel Euismod Eleifend Pharetra George Himenaeos Facilisis Conubia Phasellus Ipsum Turpis Varius Cursus Suscipit Inceptos Nulla Per Varius 
		Auctor Non Habitant Pretium Mollis Conubia Sodales Congue Laoreet Etiam Mollis Congue Lorem Tempor Arcu Lectus Sodales Euismod Nibh Orci Lectus Aenean Ultricies Donec Massa Quam Vehicula Fames Dictumst Scelerisque Vitae Erat Odio Vulputate Suscipit Facilisis Lacus Habitasse A Odio Consequat Aliquet Non Commodo Dui Feug
		iat Duis Nibh Eget Viverra Dictum Maecenas Semper Odio Donec Donec Et Eleifend Quis Morbi Lorem Nostra Curabitur Aliquam Hendrerit Dapibus Viverra Sociosqu Lorem Consequat 
		Non Massa Imperdiet Porta Consequat Praesent Sit Consectetur Accumsan Proin Et Pellentesque Pellentesque Sollicitudin Mollis Sapien Vulputate Quisque Accumsan Pulvinar Lacinia Vulputate 
		Mauris Congue Arcu At Condimentum Metus Quis Vehicula Himenaeos Ipsum Id Fermentum Ullamcorper Mauris Justo Venenatis Ac Netus Justo Himenaeos Morbi Torquent Aptent Varius Nibh Suspendisse 
		Pellentesque Et Turpis Porta Sociosqu Tempor Egestas Orci Elementum 
		Curabitur Morbi Himenaeos Suscipit Lectus Porta Varius Suspendisse Habitant Nulla Auctor Tincidunt Dapibus Fermentum Ornare Cubilia Ac Fringilla Mattis Etiam Facilisis 
		Nullam Phasellus Sem Odio Vehicula Massa Rutrum Quisque Nec Urna Sagittis Platea Nibh Pulvinar Risus Laoreet Duis Sit Vulputate Vestibulum Justo Et Interdum Laoreet Augue Mollis 
		Rutrum Egestas Fermentum Tempor Convallis Eros Luctus Placerat Nam Aenean Vulputate Vivamus Quisque Nostra Semper Facilisis Platea Ac Hac Nam Ipsum Lobortis Mollis Hac Enim Leo Phasellus Netus Himenaeos Ornare Per Rutrum Neque Nam Congue Non Odio Donec Vehicula Est Morbi Curabitur 
		Torquent Non Lorem Cubilia At Dolor Lobortis Lectus Metus Phasellus Libero Fermentum Pulvinar Ullamcorper Quis Cursus Elit Etiam Iaculis Rhoncus Lorem Luctus 
		Tristique Himenaeos Rutrum Nam Pellentesque Et Augue Ornare Et Praesent Imperdiet Dolor Elit Posuere Habitasse Purus Iaculis Luctus Proin Fermentum Tempor Conubia Bibendum Urna Vehicula Dui Ante Nibh Fames 
		Suspendisse Sapien Mauris Est Quisque Blue Orci Est Advertisment Donec Primis Integer Risus Tristique Dapibus Vel Praesent Cursus Habitant Fames Rhoncus Magna Fermentum Eu Pharetra Molestie Eros 
		Congue Nunc Non Purus Mauris Mattis Lobortis Torquent Duis Praesent Id Est Gravida`,
		)
	})
}
