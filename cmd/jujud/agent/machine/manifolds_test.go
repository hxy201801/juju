// Copyright 2015 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package machine_test

import (
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/agent"
	"github.com/juju/juju/cmd/jujud/agent/machine"
	"github.com/juju/juju/testing"
	"github.com/juju/juju/worker"
	"github.com/juju/juju/worker/dependency"
	"github.com/juju/juju/worker/gate"
)

type ManifoldsSuite struct {
	testing.BaseSuite
}

var _ = gc.Suite(&ManifoldsSuite{})

func (s *ManifoldsSuite) TestStartFuncs(c *gc.C) {
	manifolds, _, _ := machine.Manifolds(machine.ManifoldsConfig{
		Agent: fakeAgent{},
	})
	for name, manifold := range manifolds {
		c.Logf("checking %q manifold", name)
		c.Check(manifold.Start, gc.NotNil)
	}
}

func (s *ManifoldsSuite) TestManifoldNames(c *gc.C) {
	manifolds, _, _ := machine.Manifolds(machine.ManifoldsConfig{})
	keys := make([]string, 0, len(manifolds))
	for k := range manifolds {
		keys = append(keys, k)
	}
	expectedKeys := []string{
		"agent",
		"termination",
		"api-caller",
		"upgrade-steps-gate",
		"upgrade-check-gate",
	}
	c.Assert(expectedKeys, jc.SameContents, keys)
}

func (s *ManifoldsSuite) TestUpgradeGates(c *gc.C) {
	manifolds, upgradeStepsGate, upgradeCheckGate := machine.Manifolds(machine.ManifoldsConfig{})
	assertGate(c, manifolds["upgrade-steps-gate"], upgradeStepsGate)
	assertGate(c, manifolds["upgrade-check-gate"], upgradeCheckGate)
}

func assertGate(c *gc.C, manifold dependency.Manifold, unlocker gate.Unlocker) {
	w, err := manifold.Start(nil)
	c.Assert(err, jc.ErrorIsNil)
	defer worker.Stop(w)

	var waiter gate.Waiter
	err = manifold.Output(w, &waiter)
	c.Assert(err, jc.ErrorIsNil)

	select {
	case <-waiter.Unlocked():
		c.Fatalf("expected gate to be locked")
	default:
	}

	unlocker.Unlock()

	select {
	case <-waiter.Unlocked():
	default:
		c.Fatalf("expected gate to be unlocked")
	}
}

type fakeAgent struct {
	agent.Agent
}
