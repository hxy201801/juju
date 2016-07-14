// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package featuretests

import (
	"io"
	"strings"

	"github.com/juju/cmd"
	"github.com/juju/loggo"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/juju/juju/api"
	"github.com/juju/juju/cmd/juju/commands"
	"github.com/juju/juju/juju"
	jujutesting "github.com/juju/juju/juju/testing"
	"github.com/juju/juju/testing"
)

type cmdRegistrationSuite struct {
	jujutesting.JujuConnSuite
}

func (s *cmdRegistrationSuite) run(c *gc.C, stdin io.Reader, args ...string) *cmd.Context {
	context := testing.Context(c)
	if stdin != nil {
		context.Stdin = stdin
	}
	command := commands.NewJujuCommand(context)
	c.Assert(testing.InitCommand(command, args), jc.ErrorIsNil)
	c.Assert(command.Run(context), jc.ErrorIsNil)
	loggo.RemoveWriter("warning") // remove logger added by main command
	return context
}

func (s *cmdRegistrationSuite) TestAddUserAndRegister(c *gc.C) {
	// First, add user "bob", and record the "juju register" command
	// that is printed out.
	context := s.run(c, nil, "add-user", "bob", "Bob Dobbs")
	c.Check(testing.Stderr(context), gc.Equals, "")
	stdout := testing.Stdout(context)
	c.Check(stdout, gc.Matches, `
User "Bob Dobbs \(bob\)" added
Please send this command to bob:
    juju register .*

"Bob Dobbs \(bob\)" has not been granted access to any models(.|\n)*
`[1:])
	jujuRegisterCommand := strings.Fields(strings.TrimSpace(
		strings.SplitN(stdout[strings.Index(stdout, "juju register"):], "\n", 2)[0],
	))
	c.Logf("%q", jujuRegisterCommand)

	// Now run the "juju register" command. We need to pass the
	// controller name and password to set.
	stdin := strings.NewReader("bob-controller\nhunter2\nhunter2\n")
	args := jujuRegisterCommand[1:] // drop the "juju"
	context = s.run(c, stdin, args...)
	c.Check(testing.Stdout(context), gc.Equals, "")
	c.Check(testing.Stderr(context), gc.Equals, `
WARNING: the controller proposed "kontroll" which clashes with an existing controller. The two controllers are entirely different.

Please set a name for this controller:
Enter a new password: 
Confirm password: 

Welcome, bob. You are now logged into "bob-controller".

There are no models available. You can add models with
"juju add-model", or you can ask an administrator or owner
of a model to grant access to that model with "juju grant".

`[1:])

	// Make sure that the saved server details are sufficient to connect
	// to the api server.
	accountDetails, err := s.ControllerStore.AccountDetails("bob-controller")
	c.Assert(err, jc.ErrorIsNil)
	api, err := juju.NewAPIConnection(juju.NewAPIConnectionParams{
		Store:           s.ControllerStore,
		ControllerName:  "bob-controller",
		AccountDetails:  accountDetails,
		BootstrapConfig: noBootstrapConfig,
		DialOpts:        api.DefaultDialOpts(),
		OpenAPI:         api.Open,
	})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(api.Close(), jc.ErrorIsNil)
}
