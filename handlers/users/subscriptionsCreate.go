package users

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/kataras/iris/v12"
	"github.com/ping-42/42lib/db/models"
	"github.com/ping-42/admin-api/middleware"
	"github.com/ping-42/admin-api/utils"
	"gorm.io/gorm"
)

type SubscriptionReceived struct {
	Period     time.Duration `json:"period"`
	TaskTypeId uint64        `json:"TaskTypeId"`
	TestsCount int           `json:"tests_count"` // TODO make it uint64
	Opts       interface{}   `json:"opts"`
}

type DnsOptsReceived struct {
	Host  string `json:"host" validate:"required,fqdn"`
	Proto string `json:"proto" validate:"required"`
}

type IcmpOptsReceived struct {
	Count        int      `json:"Count" validate:"required,min=1"`
	Payload      string   `json:"Payload" validate:"required"`
	TargetIPs    []string `json:"TargetIPs" validate:"dive,ip"`
	TargetDomain string   `json:"TargetDomain,omitempty"`
}

type HttpOptsReceived struct {
	HttpMethod     string              `json:"HttpMethod" validate:"required,oneof=GET POST PUT DELETE"`
	RequestBody    string              `json:"RequestBody" validate:"required"`
	TargetDomain   string              `json:"TargetDomain" validate:"required,url"`
	RequestHeaders map[string][]string `json:"RequestHeaders"`
}

type TracerouteOptsReceived struct {
	Dest       string `json:"dest" validate:"required,ip"`
	Port       int    `json:"port" validate:"required,min=1"`
	MaxHops    int    `json:"MaxHops" validate:"required,min=1"`
	Retries    int    `json:"retries" validate:"required,min=1"`
	Timeout    int    `json:"timeout" validate:"required,min=1"`
	FirstHop   int    `json:"FirstHop" validate:"required,min=1"`
	NetCapRaw  bool   `json:"NetCapRaw"`
	PacketSize int    `json:"PacketSize" validate:"required,min=1"`
}

func ServeSubscriptionsCreate(ctx iris.Context, db *gorm.DB) {

	userClaims, ok := ctx.Values().Get("UserClaims").(*middleware.UserClaims)
	if !ok {
		utils.RespondError(ctx, iris.StatusUnauthorized, "Unauthorized user", fmt.Errorf("ServeSubscriptionsCreate casting to middleware.UserClaims error"))
		return
	}

	var subscriptionReceived SubscriptionReceived
	validate := validator.New(validator.WithRequiredStructEnabled())

	if err := ctx.ReadJSON(&subscriptionReceived); err != nil {
		utils.RespondError(ctx, http.StatusBadRequest, "Invalid request", err)
		return
	}

	// Basic validation for the main fields
	if err := validateSubscription(subscriptionReceived); err != nil {
		utils.RespondError(ctx, http.StatusBadRequest, err.Error(), err)
		return
	}

	// validate opts based on TaskTypeId
	var opts []byte
	var err error

	switch subscriptionReceived.TaskTypeId {
	case 1: // DNS_TASK
		var dnsOpts DnsOptsReceived
		if err := marshalToStruct(subscriptionReceived.Opts, &dnsOpts); err != nil {
			utils.RespondError(ctx, http.StatusBadRequest, "Invalid DNS options", err)
			return
		}
		if err := validate.Struct(dnsOpts); err != nil {
			utils.RespondError(ctx, http.StatusBadRequest, "Invalid DNS options", err)
			return
		}
		opts, err = json.Marshal(dnsOpts)

	case 2: // ICMP_TASK
		var icmpOpts IcmpOptsReceived
		if err := marshalToStruct(subscriptionReceived.Opts, &icmpOpts); err != nil {
			utils.RespondError(ctx, http.StatusBadRequest, "Invalid ICMP options", err)
			return
		}
		if err := validate.Struct(icmpOpts); err != nil {
			utils.RespondError(ctx, http.StatusBadRequest, "Invalid ICMP options", err)
			return
		}

		if len(icmpOpts.TargetIPs) == 0 && icmpOpts.TargetDomain == "" {
			utils.RespondError(ctx, http.StatusBadRequest, "Invalid ICMP options, ips or domain required", err)
			return
		}

		icmpOpts.Payload = base64.StdEncoding.EncodeToString([]byte(icmpOpts.Payload))

		opts, err = json.Marshal(icmpOpts)

	case 3: // HTTP_TASK
		var httpOpts HttpOptsReceived
		if err := marshalToStruct(subscriptionReceived.Opts, &httpOpts); err != nil {
			utils.RespondError(ctx, http.StatusBadRequest, "Invalid HTTP options", err)
			return
		}
		if err := validate.Struct(httpOpts); err != nil {
			utils.RespondError(ctx, http.StatusBadRequest, "Invalid HTTP options", err)
			return
		}

		httpOpts.RequestBody = base64.StdEncoding.EncodeToString([]byte(httpOpts.RequestBody))

		opts, err = json.Marshal(httpOpts)

	case 4: // TRACEROUTE_TASK
		var tracerouteOpts TracerouteOptsReceived
		if err := marshalToStruct(subscriptionReceived.Opts, &tracerouteOpts); err != nil {
			utils.RespondError(ctx, http.StatusBadRequest, "Invalid Traceroute options", err)
			return
		}
		if err := validate.Struct(tracerouteOpts); err != nil {
			utils.RespondError(ctx, http.StatusBadRequest, "Invalid Traceroute options", err)
			return
		}
		opts, err = json.Marshal(tracerouteOpts)

	default:
		utils.RespondError(ctx, http.StatusBadRequest, "Unsupported TaskTypeId", nil)
		return
	}

	if err != nil {
		utils.RespondError(ctx, http.StatusBadRequest, "Failed to marshal opts", err)
		return
	}

	// create the new subscription
	newSubscription := models.Subscription{
		OrganizationID:       userClaims.OrganizationId,
		TaskTypeID:           subscriptionReceived.TaskTypeId,
		TestsCountSubscribed: subscriptionReceived.TestsCount,
		Period:               subscriptionReceived.Period,
		Opts:                 opts,
		IsActive:             true,
	}
	if err := db.Create(&newSubscription).Error; err != nil {
		utils.RespondError(ctx, http.StatusInternalServerError, "Failed to create subscription", err)
		return
	}

	utils.RespondCreated(ctx, newSubscription, "Subscription created successfully")
}

func validateSubscription(subscription SubscriptionReceived) error {
	if subscription.Period == 0 {
		return errors.New("Period is required")
	}
	if subscription.TaskTypeId == 0 {
		return errors.New("TaskTypeId is required")
	}
	if subscription.TestsCount == 0 {
		return errors.New("TestsCount is required")
	}
	return nil
}

func marshalToStruct(input interface{}, output interface{}) error {
	jsonData, err := json.Marshal(input)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, output)
}
