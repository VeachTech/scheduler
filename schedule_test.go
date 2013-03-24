/**
Copyright (c) 2013, Ryan Veach
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:
    * Redistributions of source code must retain the above copyright
      notice, this list of conditions and the following disclaimer.
    * Redistributions in binary form must reproduce the above copyright
      notice, this list of conditions and the following disclaimer in the
      documentation and/or other materials provided with the distribution.
    * Neither the name of the <organization> nor the
      names of its contributors may be used to endorse or promote products
      derived from this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL <COPYRIGHT HOLDER> BE LIABLE FOR ANY
DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
(INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
**/

package scheduler

import (
	"testing"
	"time"
)

func TestID(t *testing.T) {
	var id ID
	if id.GetID() != 0 {
		t.Errorf("new default ID = %v, should be %v", id, 0)
	}

	var in, out uint = 1, 1
	id = ID(in)
	if id.GetID() != out {
		t.Errorf("ID(%v) = %v, want %v", in, id, out)
	}

	id = 3
	if id.GetID() != 3 {
		t.Errorf("id = %v", id.GetID())
	}
}

func TestCron(t *testing.T) {
	check, err := validMatch("*/2", int(time.Month(1))-1)
	if !check || err != nil {
		t.Errorf("validMonth(\"*/2\", time.Month(1)) resulted in %v with error %v", check, err)
	}

	check, err = validMatch("*/2", int(time.Month(7))-1)
	if !check || err != nil {
		t.Errorf("validMonth(\"*/2\", time.Month(7)) resulted in %v with error %v", check, err)
	}

	check, err = validMatch("*/3", int(time.Month(4))-1)
	if !check || err != nil {
		t.Errorf("validMonth(\"*/3\", time.Month(4)) resulted in %v with error %v", check, err)
	}

	check, err = validMatch("*/3", int(time.Month(3))-1)
	if check || err != nil {
		t.Errorf("validMonth(\"*/3\", time.Month(3)) resulted in %v with error %v", check, err)
	}

	check, err = validMatch("1-3", int(time.Month(3))-1)
	if !check || err != nil {
		t.Errorf("validMonth(\"1-3\", time.Month(3)) resulted in %v with error %v", check, err)
	}
}
