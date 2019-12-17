package coach

import (
	"context"
	"github.com/golang/mock/gomock"
	"github.com/lugobots/lugo4go/v2"
	"github.com/lugobots/lugo4go/v2/proto"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDefineMyState_AllStates(t *testing.T) {
	var state PlayerState
	var err error
	home3 := &proto.Player{Number: 3, TeamSide: proto.Team_HOME}
	home5 := &proto.Player{Number: 5, TeamSide: proto.Team_HOME}
	away5 := &proto.Player{Number: 5, TeamSide: proto.Team_AWAY}
	ball := &proto.Ball{}

	snapshot := &proto.GameSnapshot{
		Ball:     ball,
		HomeTeam: &proto.Team{Players: []*proto.Player{home3, home5}},
		AwayTeam: &proto.Team{Players: []*proto.Player{away5}},
	}

	// everyone is disputing the ball
	state, err = DefineMyState(lugo4go.Config{Number: 3, TeamSide: proto.Team_HOME}, snapshot)
	assert.Nil(t, err)
	assert.Equal(t, DisputingTheBall, state)

	state, err = DefineMyState(lugo4go.Config{Number: 5, TeamSide: proto.Team_HOME}, snapshot)
	assert.Nil(t, err)
	assert.Equal(t, DisputingTheBall, state)

	state, err = DefineMyState(lugo4go.Config{Number: 5, TeamSide: proto.Team_AWAY}, snapshot)
	assert.Nil(t, err)
	assert.Equal(t, DisputingTheBall, state)

	ball.Holder = home3

	// Holding
	state, err = DefineMyState(lugo4go.Config{Number: 3, TeamSide: proto.Team_HOME}, snapshot)
	assert.Nil(t, err)
	assert.Equal(t, HoldingTheBall, state)

	// supporting
	state, err = DefineMyState(lugo4go.Config{Number: 5, TeamSide: proto.Team_HOME}, snapshot)
	assert.Nil(t, err)
	assert.Equal(t, Supporting, state)

	//
	state, err = DefineMyState(lugo4go.Config{Number: 5, TeamSide: proto.Team_AWAY}, snapshot)
	assert.Nil(t, err)
	assert.Equal(t, Defending, state)

}

func TestDefineMyState_ErrorInvalidSnapshot(t *testing.T) {
	var err error

	_, err = DefineMyState(lugo4go.Config{Number: 3, TeamSide: proto.Team_HOME}, nil)
	assert.Equal(t, err, ErrNoBall)

	_, err = DefineMyState(lugo4go.Config{Number: 3, TeamSide: proto.Team_HOME}, &proto.GameSnapshot{})
	assert.Equal(t, err, ErrNoBall)
}

func TestDefineMyState_ErrorNoPlayer(t *testing.T) {
	var err error

	_, err = DefineMyState(lugo4go.Config{Number: 3, TeamSide: proto.Team_HOME}, &proto.GameSnapshot{Ball: &proto.Ball{}})
	assert.Equal(t, err, ErrPlayerNotFound)
}

func TestDefaultTurnHandler_ShouldCallRightMethod(t *testing.T) {
	// initiates Mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // checks all expected things for mocks

	mockDecider := NewMockDecider(ctrl)
	mockSender := lugo4go.NewMockOrderSender(ctrl)

	config := lugo4go.Config{Number: 4, TeamSide: proto.Team_HOME}

	defaultHandler := DefaultTurnHandler(mockDecider, config, nil)

	ctx, stop := context.WithTimeout(context.Background(), 1*time.Second)
	defer stop()

	me := &proto.Player{Number: config.Number, TeamSide: config.TeamSide}
	myFriend := &proto.Player{Number: 5, TeamSide: config.TeamSide}
	myOpponent := &proto.Player{Number: 5, TeamSide: proto.Team_AWAY}

	ball := &proto.Ball{}
	snapshot := &proto.GameSnapshot{
		Ball:     ball,
		HomeTeam: &proto.Team{Players: []*proto.Player{me}},
	}

	// disputing
	expectedTurnData := TurnData{Me: me, Sender: mockSender, Snapshot: snapshot}
	mockDecider.EXPECT().OnDisputing(ctx, expectedTurnData)
	defaultHandler(ctx, snapshot, mockSender)

	// disputing
	ball.Holder = myFriend
	expectedTurnData = TurnData{Me: me, Sender: mockSender, Snapshot: snapshot}
	mockDecider.EXPECT().OnSupporting(ctx, expectedTurnData)
	defaultHandler(ctx, snapshot, mockSender)

	// Defending
	ball.Holder = myOpponent
	expectedTurnData = TurnData{Me: me, Sender: mockSender, Snapshot: snapshot}
	mockDecider.EXPECT().OnDefending(ctx, expectedTurnData)
	defaultHandler(ctx, snapshot, mockSender)

	// holding
	ball.Holder = me
	expectedTurnData = TurnData{Me: me, Sender: mockSender, Snapshot: snapshot}
	mockDecider.EXPECT().OnHolding(ctx, expectedTurnData)
	defaultHandler(ctx, snapshot, mockSender)

}

func TestDefaultTurnHandler_ShouldCallGoalkeeperMethod(t *testing.T) {
	// initiates Mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // checks all expected things for mocks

	mockDecider := NewMockDecider(ctrl)
	mockSender := lugo4go.NewMockOrderSender(ctrl)

	config := lugo4go.Config{Number: 1, TeamSide: proto.Team_HOME}

	defaultHandler := DefaultTurnHandler(mockDecider, config, nil)

	ctx, stop := context.WithTimeout(context.Background(), 1*time.Second)
	defer stop()

	me := &proto.Player{Number: config.Number, TeamSide: config.TeamSide}
	myFriend := &proto.Player{Number: 5, TeamSide: config.TeamSide}
	myOpponent := &proto.Player{Number: 5, TeamSide: proto.Team_AWAY}

	ball := &proto.Ball{}
	snapshot := &proto.GameSnapshot{
		Ball:     ball,
		HomeTeam: &proto.Team{Players: []*proto.Player{me}},
	}

	// disputing
	expectedTurnData := TurnData{Me: me, Sender: mockSender, Snapshot: snapshot}
	mockDecider.EXPECT().AsGoalkeeper(ctx, expectedTurnData)
	defaultHandler(ctx, snapshot, mockSender)

	// disputing
	ball.Holder = myFriend
	expectedTurnData = TurnData{Me: me, Sender: mockSender, Snapshot: snapshot}
	mockDecider.EXPECT().AsGoalkeeper(ctx, expectedTurnData)
	defaultHandler(ctx, snapshot, mockSender)

	// Defending
	ball.Holder = myOpponent
	expectedTurnData = TurnData{Me: me, Sender: mockSender, Snapshot: snapshot}
	mockDecider.EXPECT().AsGoalkeeper(ctx, expectedTurnData)
	defaultHandler(ctx, snapshot, mockSender)

	// holding
	ball.Holder = me
	expectedTurnData = TurnData{Me: me, Sender: mockSender, Snapshot: snapshot}
	mockDecider.EXPECT().AsGoalkeeper(ctx, expectedTurnData)
	defaultHandler(ctx, snapshot, mockSender)

}

func TestDefaultTurnHandler_ShouldPanicIfPlayerIsNotThere(t *testing.T) {
	// initiates Mock controller
	ctrl := gomock.NewController(t)
	defer ctrl.Finish() // checks all expected things for mocks

	mockDecider := NewMockDecider(ctrl)
	mockSender := lugo4go.NewMockOrderSender(ctrl)

	config := lugo4go.Config{Number: 1, TeamSide: proto.Team_HOME}

	defaultHandler := DefaultTurnHandler(mockDecider, config, nil)

	ctx, stop := context.WithTimeout(context.Background(), 1*time.Second)
	defer stop()

	ball := &proto.Ball{}
	snapshot := &proto.GameSnapshot{
		Ball:     ball,
		HomeTeam: &proto.Team{Players: []*proto.Player{}},
	}
	assert.PanicsWithValue(t, "i did not find my self in the game", func() {
		defaultHandler(ctx, snapshot, mockSender)
	})
}
