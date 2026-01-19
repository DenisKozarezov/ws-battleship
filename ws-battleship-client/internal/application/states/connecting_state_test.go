package states

import (
	"context"
	"errors"
	"testing"
	"time"
	client "ws-battleship-client/internal/delivery/websocket"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestConnectingStateErrorCallbacks(t *testing.T) {
	for _, tt := range []struct {
		name    string
		prepare func(m *client.MockClient)
		err     error
	}{
		{
			name: "some error while connecting to server",
			prepare: func(m *client.MockClient) {
				m.On("Connect", mock.Anything, mock.Anything).Return(errors.New("error"))
			},
			err: errors.New("error"),
		},
		{
			name: "deadline exceeded while connecting to server",
			prepare: func(m *client.MockClient) {
				m.On("Connect", mock.Anything, mock.Anything).Return(context.DeadlineExceeded)
			},
			err: context.DeadlineExceeded,
		},
		{
			name: "if no error while connecting to server, then no error callback called",
			prepare: func(m *client.MockClient) {
				m.On("Connect", mock.Anything, mock.Anything).Return(nil)
			},
			err: nil,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Arrange
			var callbackErr error
			callback := func(err error) {
				callbackErr = err
			}
			stateMachine := NewStateMachine()

			clientMock := client.NewMockClient(t)
			if tt.prepare != nil {
				tt.prepare(clientMock)
			}

			// 2. Act
			state := &ConnectingState{
				stateMachine: stateMachine,
				client:       clientMock,
				onError:      callback,
			}
			state.OnEnter()

			// 3. Assert
			time.Sleep(time.Millisecond * 100)
			if tt.err != nil {
				require.Error(t, callbackErr)
				require.ErrorContains(t, tt.err, callbackErr.Error())
			} else {
				require.NoError(t, callbackErr)
			}
		})
	}
}

func TestConnectingStateSuccessCallbacks(t *testing.T) {
	for _, tt := range []struct {
		name           string
		prepare        func(m *client.MockClient)
		callbackCalled bool
	}{
		{
			name: "some error while connecting to server",
			prepare: func(m *client.MockClient) {
				m.On("Connect", mock.Anything, mock.Anything).Return(errors.New("error"))
			},
			callbackCalled: false,
		},
		{
			name: "if no error while connecting to server, then success callback called",
			prepare: func(m *client.MockClient) {
				m.On("Connect", mock.Anything, mock.Anything).Return(nil)
			},
			callbackCalled: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// 1. Arrange
			var callbackCalled bool
			callback := func(client client.Client) {
				callbackCalled = true
			}
			stateMachine := NewStateMachine()

			clientMock := client.NewMockClient(t)
			if tt.prepare != nil {
				tt.prepare(clientMock)
			}

			// 2. Act
			state := &ConnectingState{
				stateMachine: stateMachine,
				client:       clientMock,
				onSuccess:    callback,
			}
			state.OnEnter()

			// 3. Assert
			time.Sleep(time.Millisecond * 100)
			require.Equal(t, tt.callbackCalled, callbackCalled)
		})
	}
}
