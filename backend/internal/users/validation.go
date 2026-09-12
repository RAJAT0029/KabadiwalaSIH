package users

import (
	"fmt"
	"kabadiconnect/backend/internal/materials"
	"math"
	"strings"
)

func ValidateProfilePatch(in ProfilePatch) error {
	for _, v := range []*string{in.Name, in.CompanyName, in.FacilityName} {
		if v != nil && (strings.TrimSpace(*v) == "" || len(*v) > 200) {
			return fmt.Errorf("name and facility fields must contain 1–200 bytes")
		}
	}
	if in.AcceptedMaterials != nil {
		if len(*in.AcceptedMaterials) > len(materials.Categories) {
			return fmt.Errorf("too many accepted materials")
		}
		for _, c := range *in.AcceptedMaterials {
			if !materials.Supported(c) {
				return fmt.Errorf("unsupported e-waste material category")
			}
		}
	}
	if in.ProcessingCapacityKg != nil && (math.IsNaN(*in.ProcessingCapacityKg) || math.IsInf(*in.ProcessingCapacityKg, 0) || *in.ProcessingCapacityKg < 0) {
		return fmt.Errorf("processing capacity must be finite and non-negative")
	}
	if in.Authorization != nil && (in.Authorization.Status != "" || in.Authorization.IsDemo || in.Authorization.ValidFrom != nil || in.Authorization.ValidUntil != nil) {
		return fmt.Errorf("authorization status, demo flag and validity dates are verification-service fields")
	}
	if in.OfferedRates != nil {
		for _, rate := range *in.OfferedRates {
			if !materials.Supported(rate.MaterialCategory) || rate.PricePerUnit <= 0 || math.IsNaN(rate.PricePerUnit) || math.IsInf(rate.PricePerUnit, 0) || (rate.Unit != "kg" && rate.Unit != "unit") {
				return fmt.Errorf("invalid material rate")
			}
		}
	}
	return nil
}
