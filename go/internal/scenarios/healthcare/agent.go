package healthcare

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// -- State --

type AgentState struct {
	Patient  *Patient          `json:"patient"`
	Messages []*schema.Message `json:"messages"`
}

type Patient struct {
	PatientID string `json:"patient_id"`
	Name      string `json:"name"`
	Age       int    `json:"age,omitempty"`
	Insurance string `json:"insurance,omitempty"`
	Status    string `json:"status,omitempty"`
}

// -- Tool Args --

type AssessSymptomsArgs struct {
	PatientID string   `json:"patient_id" desc:"Patient ID"`
	Symptoms  []string `json:"symptoms" desc:"List of symptoms"`
	Urgency   string   `json:"urgency" desc:"Urgency level"`
}

type RegisterPatientArgs struct {
	Name              string `json:"name" desc:"Patient Name"`
	DateOfBirth       string `json:"date_of_birth" desc:"DOB"`
	InsuranceProvider string `json:"insurance_provider" desc:"Insurance Provider"`
}

type ScheduleAppointmentArgs struct {
	PatientID       string `json:"patient_id" desc:"Patient ID"`
	AppointmentType string `json:"appointment_type" desc:"Type of appointment"`
	Provider        string `json:"provider" desc:"Healthcare provider"`
}

type VerifyInsuranceArgs struct {
	PatientID         string `json:"patient_id" desc:"Patient ID"`
	InsuranceProvider string `json:"insurance_provider" desc:"Insurance Provider"`
	PolicyNumber      string `json:"policy_number" desc:"Policy Number"`
}

type UpdateMedicalHistoryArgs struct {
	PatientID string `json:"patient_id" desc:"Patient ID"`
}

type ReferSpecialistArgs struct {
	PatientID string `json:"patient_id" desc:"Patient ID"`
	Specialty string `json:"specialty" desc:"Specialty"`
	Reason    string `json:"reason" desc:"Reason for referral"`
}

type PrescribeMedicationArgs struct {
	PatientID  string `json:"patient_id" desc:"Patient ID"`
	Medication string `json:"medication" desc:"Medication name"`
	Dosage     string `json:"dosage" desc:"Dosage instructions"`
}

type SendPatientMessageArgs struct {
	PatientID string `json:"patient_id" desc:"Patient ID"`
	Message   string `json:"message" desc:"Message content"`
}

// -- Tool Impls --

func AssessSymptoms(ctx context.Context, args *AssessSymptomsArgs) (string, error) {
	fmt.Printf("[TOOL] assess_symptoms(pat=%s, urgency=%s)\n", args.PatientID, args.Urgency)
	return "symptoms_assessed", nil
}

func RegisterPatient(ctx context.Context, args *RegisterPatientArgs) (string, error) {
	fmt.Printf("[TOOL] register_patient(name=%s, provider=%s)\n", args.Name, args.InsuranceProvider)
	return "patient_registered", nil
}

func ScheduleAppointment(ctx context.Context, args *ScheduleAppointmentArgs) (string, error) {
	fmt.Printf("[TOOL] schedule_appointment(pat=%s, type=%s)\n", args.PatientID, args.AppointmentType)
	return "appointment_scheduled", nil
}

func VerifyInsurance(ctx context.Context, args *VerifyInsuranceArgs) (string, error) {
	fmt.Printf("[TOOL] verify_insurance(pat=%s, provider=%s)\n", args.PatientID, args.InsuranceProvider)
	return "insurance_verified", nil
}

func UpdateMedicalHistory(ctx context.Context, args *UpdateMedicalHistoryArgs) (string, error) {
	fmt.Printf("[TOOL] update_medical_history(pat=%s)\n", args.PatientID)
	return "medical_history_updated", nil
}

func ReferSpecialist(ctx context.Context, args *ReferSpecialistArgs) (string, error) {
	fmt.Printf("[TOOL] refer_specialist(pat=%s, spec=%s)\n", args.PatientID, args.Specialty)
	return "referral_created", nil
}

func PrescribeMedication(ctx context.Context, args *PrescribeMedicationArgs) (string, error) {
	fmt.Printf("[TOOL] prescribe_medication(pat=%s, med=%s)\n", args.PatientID, args.Medication)
	return "prescription_sent", nil
}

func SendPatientMessage(ctx context.Context, args *SendPatientMessageArgs) (string, error) {
	fmt.Printf("[TOOL] send_patient_message -> %s\n", args.Message)
	return "message_sent", nil
}

// -- Graph --

func NewAgent(ctx context.Context) (compose.Runnable[*AgentState, *AgentState], error) {
	temp := float32(0.0)
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		Model:       "gpt-4o",
		Temperature: &temp,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to init chat model: %v", err)
	}

	strParam := func(desc string) *schema.ParameterInfo {
		return &schema.ParameterInfo{Type: schema.String, Desc: desc, Required: true}
	}
	strParamOpt := func(desc string) *schema.ParameterInfo {
		return &schema.ParameterInfo{Type: schema.String, Desc: desc, Required: false}
	}
	arrParam := func(desc string) *schema.ParameterInfo {
		return &schema.ParameterInfo{Type: schema.Array, Desc: desc, Required: false, ElemInfo: &schema.ParameterInfo{Type: schema.String}} // Assuming array of strings
	}

	tools := []*schema.ToolInfo{
		{
			Name: "assess_symptoms",
			Desc: "Assess patient symptoms and determine urgency level for triage.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"patient_id": strParam("Patient ID"),
				"symptoms":   arrParam("List of symptoms"),
				"urgency":    strParamOpt("Urgency level"),
			}),
		},
		{
			Name: "register_patient",
			Desc: "Register a new patient in the healthcare system.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"name":               strParam("Patient Name"),
				"date_of_birth":      strParamOpt("DOB"),
				"insurance_provider": strParamOpt("Insurance Provider"),
			}),
		},
		{
			Name: "schedule_appointment",
			Desc: "Schedule appointments for patients with healthcare providers.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"patient_id":       strParam("Patient ID"),
				"appointment_type": strParam("Type of appointment"),
				"provider":         strParamOpt("Healthcare provider"),
			}),
		},
		{
			Name: "verify_insurance",
			Desc: "Verify patient insurance coverage and eligibility.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"patient_id":         strParam("Patient ID"),
				"insurance_provider": strParam("Insurance Provider"),
				"policy_number":      strParamOpt("Policy Number"),
			}),
		},
		{
			Name: "update_medical_history",
			Desc: "Update patient medical history including medications, allergies, and family history.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"patient_id": strParam("Patient ID"),
			}),
		},
		{
			Name: "refer_specialist",
			Desc: "Refer patient to specialist for specialized care.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"patient_id": strParam("Patient ID"),
				"specialty":  strParam("Specialty"),
				"reason":     strParamOpt("Reason for referral"),
			}),
		},
		{
			Name: "prescribe_medication",
			Desc: "Prescribe or refill medications for patients.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"patient_id": strParam("Patient ID"),
				"medication": strParam("Medication name"),
				"dosage":     strParamOpt("Dosage instructions"),
			}),
		},
		{
			Name: "send_patient_message",
			Desc: "Send a message or response to the patient.",
			ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
				"patient_id": strParam("Patient ID"),
				"message":    strParam("Message content"),
			}),
		},
	}

	if err := chatModel.BindTools(tools); err != nil {
		return nil, err
	}

	toolHandlers := map[string]func(ctx context.Context, args interface{}) (string, error){
		"assess_symptoms": func(ctx context.Context, args interface{}) (string, error) {
			return AssessSymptoms(ctx, args.(*AssessSymptomsArgs))
		},
		"register_patient": func(ctx context.Context, args interface{}) (string, error) {
			return RegisterPatient(ctx, args.(*RegisterPatientArgs))
		},
		"schedule_appointment": func(ctx context.Context, args interface{}) (string, error) {
			return ScheduleAppointment(ctx, args.(*ScheduleAppointmentArgs))
		},
		"verify_insurance": func(ctx context.Context, args interface{}) (string, error) {
			return VerifyInsurance(ctx, args.(*VerifyInsuranceArgs))
		},
		"update_medical_history": func(ctx context.Context, args interface{}) (string, error) {
			return UpdateMedicalHistory(ctx, args.(*UpdateMedicalHistoryArgs))
		},
		"refer_specialist": func(ctx context.Context, args interface{}) (string, error) {
			return ReferSpecialist(ctx, args.(*ReferSpecialistArgs))
		},
		"prescribe_medication": func(ctx context.Context, args interface{}) (string, error) {
			return PrescribeMedication(ctx, args.(*PrescribeMedicationArgs))
		},
		"send_patient_message": func(ctx context.Context, args interface{}) (string, error) {
			return SendPatientMessage(ctx, args.(*SendPatientMessageArgs))
		},
	}

	assistant := func(ctx context.Context, state *AgentState) (*AgentState, error) {
		patientJSON, _ := json.Marshal(state.Patient)
		sysPrompt := fmt.Sprintf(
			"You are a professional healthcare patient intake and triage specialist.\n"+
				"Your role is to help patients with: Registration, Triage, Appointments, History, Referrals.\n"+
				"When assisting patients:\n"+
				"  1) Call the appropriate healthcare tool based on their needs\n"+
				"  2) Follow up with send_patient_message to confirm actions taken\n"+
				"Always prioritize patient safety and ensure urgent cases are handled immediately.\n\n"+
				"PATIENT: %s", string(patientJSON))

		inputMsgs := append([]*schema.Message{schema.SystemMessage(sysPrompt)}, state.Messages...)
		resp, err := chatModel.Generate(ctx, inputMsgs)
		if err != nil {
			return nil, err
		}
		state.Messages = append(state.Messages, resp)
		return state, nil
	}

	toolExecutor := func(ctx context.Context, state *AgentState) (*AgentState, error) {
		lastMsg := state.Messages[len(state.Messages)-1]
		if len(lastMsg.ToolCalls) == 0 {
			return state, nil
		}

		for _, tc := range lastMsg.ToolCalls {
			handler, ok := toolHandlers[tc.Function.Name]
			if !ok {
				log.Printf("Tool %s not found", tc.Function.Name)
				continue
			}

			var resultStr string
			var err error

			switch tc.Function.Name {
			case "assess_symptoms":
				var a AssessSymptomsArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "register_patient":
				var a RegisterPatientArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "schedule_appointment":
				var a ScheduleAppointmentArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "verify_insurance":
				var a VerifyInsuranceArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "update_medical_history":
				var a UpdateMedicalHistoryArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "refer_specialist":
				var a ReferSpecialistArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "prescribe_medication":
				var a PrescribeMedicationArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			case "send_patient_message":
				var a SendPatientMessageArgs
				_ = json.Unmarshal([]byte(tc.Function.Arguments), &a)
				resultStr, err = handler(ctx, &a)
			}

			if err != nil {
				resultStr = fmt.Sprintf("Error: %v", err)
			}
			state.Messages = append(state.Messages, &schema.Message{
				Role:       schema.Tool,
				Content:    resultStr,
				ToolCallID: tc.ID,
			})
		}
		return state, nil
	}

	g := compose.NewGraph[*AgentState, *AgentState]()
	_ = g.AddLambdaNode("assistant", compose.InvokableLambda(assistant))
	_ = g.AddLambdaNode("tools", compose.InvokableLambda(toolExecutor))
	_ = g.AddEdge(compose.START, "assistant")

	_ = g.AddBranch("assistant", compose.NewGraphBranch(func(_ context.Context, state *AgentState) (string, error) {
		lastMsg := state.Messages[len(state.Messages)-1]
		if len(lastMsg.ToolCalls) > 0 {
			return "tools", nil
		}
		return compose.END, nil
	}, map[string]bool{"tools": true, compose.END: true}))

	_ = g.AddEdge("tools", "assistant")

	return g.Compile(ctx)
}
