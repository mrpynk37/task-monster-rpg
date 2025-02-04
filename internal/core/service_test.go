package core

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"rpgMonster/internal/model"
)

func TestNewService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	dbClient := NewMockDBClient(ctrl)
	gptClient := NewMockGPTClient(ctrl)

	type args struct {
		gptClient GPTClient
		dbClient  DBClient
	}
	tests := []struct {
		name string
		args args
		want *Service
	}{
		{
			name: "normal_case",
			args: args{
				gptClient: gptClient,
				dbClient:  dbClient,
			},
			want: &Service{
				gptClient: gptClient,
				dbManager: dbClient,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := NewService(tt.args.gptClient, tt.args.dbClient)
			require.Equal(t, tt.want, res)
		})
	}
}

func TestService_CreateTaskFromGPTByRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	dbClient := NewMockDBClient(ctrl)
	gptClient := NewMockGPTClient(ctrl)
	svc := NewService(gptClient, dbClient)

	type args struct {
		req    string
		userID string
	}
	tests := []struct {
		name     string
		args     args
		mockFunc func(db *MockDBClient, gpt *MockGPTClient)
		wantErr  bool
		expRes   *model.Task
	}{
		{
			name: "empty_request",
			args: args{
				req:    "",
				userID: "",
			},
			wantErr: true,
		},
		{
			name: "normal_case",
			args: args{
				req:    "C#",
				userID: "1",
			},
			wantErr: false,
			mockFunc: func(db *MockDBClient, gpt *MockGPTClient) {
				req := "Write a one single daily task to achieve goal learn C#, in format: 'daily task: task description: requirements to check' and delimiter is comma"
				gpt.EXPECT().GetCompletion(model.GPT_SYSTEM_PROMPT, req).Return(model.GPTAnswer{
					Choices: []model.GPTChoice{
						{
							Message: model.GPTMessage{
								Content: "learn C# and test description",
							},
						},
					},
				}, nil)
				db.EXPECT().CreateTask(gomock.Any(), gomock.Any()).Return(nil)
			},
			expRes: &model.Task{
				Title:       "learn C#",
				Description: "learn C# and test description",
				Executor:    "1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockFunc != nil {
				tt.mockFunc(dbClient, gptClient)
			}
			res, err := svc.CreateTaskFromGPTByRequest(tt.args.req, tt.args.userID)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, res)
				require.Equal(t, tt.expRes, res)
			}
		})
	}
}
