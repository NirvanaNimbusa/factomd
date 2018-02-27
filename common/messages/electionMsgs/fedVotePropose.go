// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package electionMsgs

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	log "github.com/sirupsen/logrus"
	//"github.com/FactomProject/factomd/state"
)

var _ = fmt.Print

// FedVoteProposalMsg is not a majority, it is just proposing a volunteer
type FedVoteProposalMsg struct {
	FedVoteMsg

	// Signer of the message
	Signer interfaces.IHash

	// Volunteer fields
	Volunteer FedVoteVolunteerMsg

	messageHash interfaces.IHash
}

var _ interfaces.IMsg = (*FedVoteVolunteerMsg)(nil)
var _ interfaces.IElectionMsg = (*FedVoteVolunteerMsg)(nil)

func NewFedProposalMsg(signer interfaces.IHash, vol FedVoteVolunteerMsg) *FedVoteProposalMsg {
	p := new(FedVoteProposalMsg)
	p.Volunteer = vol
	p.Signer = signer

	return p
}

func (m *FedVoteProposalMsg) ElectionProcess(is interfaces.IState, elect interfaces.IElections) {

}

var _ interfaces.IMsg = (*FedVoteVolunteerMsg)(nil)

func (a *FedVoteProposalMsg) IsSameAs(msg interfaces.IMsg) bool {

	return false
}

func (m *FedVoteProposalMsg) GetServerID() interfaces.IHash {
	return m.Signer
}

func (m *FedVoteProposalMsg) LogFields() log.Fields {
	return log.Fields{"category": "message", "messagetype": "FedVoteVolunteerMsg", "dbheight": m.DBHeight, "newleader": m.Signer.String()[4:12]}
}

func (m *FedVoteProposalMsg) GetRepeatHash() interfaces.IHash {
	return m.GetMsgHash()
}

// We have to return the haswh of the underlying message.

func (m *FedVoteProposalMsg) GetHash() interfaces.IHash {
	return m.GetMsgHash()
}

func (m *FedVoteProposalMsg) GetTimestamp() interfaces.Timestamp {
	return m.TS
}

func (m *FedVoteProposalMsg) GetMsgHash() interfaces.IHash {
	if m.MsgHash == nil {
		data, err := m.MarshalBinary()
		if err != nil {
			return nil
		}
		m.MsgHash = primitives.Sha(data)
	}
	return m.MsgHash
}

func (m *FedVoteProposalMsg) Type() byte {
	return constants.VOLUNTEERAUDIT
}

func (m *FedVoteProposalMsg) Validate(state interfaces.IState) int {
	return 1
}

// Returns true if this is a message for this server to execute as
// a leader.
func (m *FedVoteProposalMsg) ComputeVMIndex(state interfaces.IState) {
}

// Execute the leader functions of the given message
// Leader, follower, do the same thing.
func (m *FedVoteProposalMsg) LeaderExecute(state interfaces.IState) {
	m.FollowerExecute(state)
}

func (m *FedVoteProposalMsg) FollowerExecute(state interfaces.IState) {
	state.ElectionsQueue().Enqueue(m)
}

// Acknowledgements do not go into the process list.
func (e *FedVoteProposalMsg) Process(dbheight uint32, state interfaces.IState) bool {
	panic("Election object should never have its Process() method called")
}

func (e *FedVoteProposalMsg) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *FedVoteProposalMsg) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (m *FedVoteProposalMsg) UnmarshalBinaryData(data []byte) (newData []byte, err error) {

	return
}

func (m *FedVoteProposalMsg) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (m *FedVoteProposalMsg) MarshalBinary() (data []byte, err error) {
	var buf primitives.Buffer

	data = buf.DeepCopyBytes()
	return data, nil
}

func (m *FedVoteProposalMsg) String() string {
	return ""
}
