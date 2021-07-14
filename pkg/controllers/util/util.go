/*
Copyright 2021 The Alibaba Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package util

import (
	"errors"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	"github.com/alibaba/morphling/pkg/controllers/consts"
)

func getConditionExperiment(exp *morphlingv1alpha1.ProfilingExperiment, condType morphlingv1alpha1.ProfilingConditionType) *morphlingv1alpha1.ProfilingCondition {
	if exp.Status.Conditions != nil {
		for _, condition := range exp.Status.Conditions {
			if condition.Type == condType {
				return &condition
			}
		}
	}
	return nil
}

func hasConditionExperiment(exp *morphlingv1alpha1.ProfilingExperiment, condType morphlingv1alpha1.ProfilingConditionType) bool {
	cond := getConditionExperiment(exp, condType)
	if cond != nil && cond.Status == v1.ConditionTrue {
		return true
	}
	return false
}
func GetLastConditionTypeProfiling(exp *morphlingv1alpha1.ProfilingExperiment) (morphlingv1alpha1.ProfilingConditionType, error) {
	if len(exp.Status.Conditions) > 0 {
		return exp.Status.Conditions[len(exp.Status.Conditions)-1].Type, nil
	}
	return "", errors.New("Experiment doesn't have any condition")
}

func IsFailedSampling(sampling *morphlingv1alpha1.Sampling) bool {
	return hasConditionSampling(sampling, morphlingv1alpha1.SamplingFailed)
}

func hasConditionSampling(sampling *morphlingv1alpha1.Sampling, condType morphlingv1alpha1.SamplingConditionType) bool {
	cond := getConditionSampling(sampling, condType)
	if cond != nil && cond.Status == v1.ConditionTrue {
		return true
	}
	return false
}

func IsSucceededExperiment(exp *morphlingv1alpha1.ProfilingExperiment) bool {
	return hasConditionExperiment(exp, morphlingv1alpha1.ProfilingSucceeded)
}

func IsFailedExperiment(exp *morphlingv1alpha1.ProfilingExperiment) bool {
	return hasConditionExperiment(exp, morphlingv1alpha1.ProfilingFailed)
}

func IsCompletedExperiment(exp *morphlingv1alpha1.ProfilingExperiment) bool {
	return IsSucceededExperiment(exp) || IsFailedExperiment(exp)
}

func IsCreatedExperiment(exp *morphlingv1alpha1.ProfilingExperiment) bool {
	return hasConditionExperiment(exp, morphlingv1alpha1.ProfilingCreated)
}

func MarkExperimentStatusCreated(exp *morphlingv1alpha1.ProfilingExperiment, message string) {
	setConditionExperiment(exp, morphlingv1alpha1.ProfilingCreated, v1.ConditionTrue, message)
}

func setConditionExperiment(exp *morphlingv1alpha1.ProfilingExperiment, conditionType morphlingv1alpha1.ProfilingConditionType, status v1.ConditionStatus, message string) {

	newCond := newConditionExperiment(conditionType, status, message)
	currentCond := getConditionExperiment(exp, conditionType)
	// Do nothing if condition doesn't change
	if currentCond != nil && currentCond.Status == newCond.Status {
		return
	}
	removeConditionExperiment(exp, conditionType)
	exp.Status.Conditions = append(exp.Status.Conditions, newCond)
}

func removeConditionExperiment(exp *morphlingv1alpha1.ProfilingExperiment, condType morphlingv1alpha1.ProfilingConditionType) {
	var newConditions []morphlingv1alpha1.ProfilingCondition
	for _, c := range exp.Status.Conditions {

		if c.Type == condType {
			continue
		}
		newConditions = append(newConditions, c)
	}
	exp.Status.Conditions = newConditions
}

func HasRunningTrials(exp *morphlingv1alpha1.ProfilingExperiment) bool {
	return exp.Status.TrialsRunning != 0
}

func newConditionExperiment(conditionType morphlingv1alpha1.ProfilingConditionType, status v1.ConditionStatus, message string) morphlingv1alpha1.ProfilingCondition {
	return morphlingv1alpha1.ProfilingCondition{
		Type:           conditionType,
		Status:         status,
		LastUpdateTime: metav1.Now(),
		Message:        message,
	}
}

func MarkExperimentStatusFailed(exp *morphlingv1alpha1.ProfilingExperiment, message string) {
	currentCond := getConditionExperiment(exp, morphlingv1alpha1.ProfilingRunning)
	if currentCond != nil {
		setConditionExperiment(exp, morphlingv1alpha1.ProfilingFailed, v1.ConditionFalse, currentCond.Message)
	}
	setConditionExperiment(exp, morphlingv1alpha1.ProfilingFailed, v1.ConditionTrue, message)
}

func MarkExperimentStatusSucceeded(exp *morphlingv1alpha1.ProfilingExperiment, message string) {
	currentCond := getConditionExperiment(exp, morphlingv1alpha1.ProfilingRunning)
	if currentCond != nil {
		setConditionExperiment(exp, morphlingv1alpha1.ProfilingRunning, v1.ConditionFalse, currentCond.Message)
	}
	setConditionExperiment(exp, morphlingv1alpha1.ProfilingSucceeded, v1.ConditionTrue, message)

}

func MarkExperimentStatusRunning(exp *morphlingv1alpha1.ProfilingExperiment, message string) {
	setConditionExperiment(exp, morphlingv1alpha1.ProfilingRunning, v1.ConditionTrue, message)

}

// ServicePodLabels returns the expected trial labels.
func ServiceDeploymentLabels(instance *morphlingv1alpha1.Trial) map[string]string {
	res := make(map[string]string)
	for k, v := range instance.Labels {
		res[k] = v
	}
	res["trial"] = instance.Name

	return res
}

// ServicePodLabels returns the expected trial labels.
func ServicePodLabels(instance *morphlingv1alpha1.Trial) map[string]string {
	res := make(map[string]string)
	for k, v := range instance.Labels {
		res[k] = v
	}
	res[consts.LabelTrialName] = instance.Name
	res[consts.LabelDeploymentName] = GetServiceDeploymentName(instance)

	return res
}

// ClientLabels returns the expected trial labels.
func ClientLabels(instance *morphlingv1alpha1.Trial) map[string]string {
	res := make(map[string]string)
	for k, v := range instance.Labels {
		res[k] = v
	}
	res["trial"] = instance.Name
	return res
}

// Trial related

func IsCreatedTrial(trial *morphlingv1alpha1.Trial) bool {
	return hasConditionTrial(trial, morphlingv1alpha1.TrialCreated)
}

func hasConditionTrial(trial *morphlingv1alpha1.Trial, condType morphlingv1alpha1.TrialConditionType) bool {
	cond := getConditionTrial(trial, condType)
	if cond != nil && cond.Status == v1.ConditionTrue {
		return true
	}
	return false
}

func getConditionTrial(trial *morphlingv1alpha1.Trial, condType morphlingv1alpha1.TrialConditionType) *morphlingv1alpha1.TrialCondition {
	for _, condition := range trial.Status.Conditions {
		if condition.Type == condType {
			return &condition
		}
	}
	return nil
}

func newConditionTrial(conditionType morphlingv1alpha1.TrialConditionType, status v1.ConditionStatus, message string) morphlingv1alpha1.TrialCondition {
	return morphlingv1alpha1.TrialCondition{
		Type:               conditionType,
		Status:             status,
		LastUpdateTime:     metav1.Now(),
		Message:            message,
		LastTransitionTime: metav1.Now(),
	}
}

func SetConditionTrial(trial *morphlingv1alpha1.Trial, conditionType morphlingv1alpha1.TrialConditionType, status v1.ConditionStatus, message string) {

	newCond := newConditionTrial(conditionType, status, message)
	currentCond := getConditionTrial(trial, conditionType)
	// Do nothing if condition doesn't change
	if currentCond != nil && currentCond.Status == newCond.Status {
		return
	}

	// Do not update lastTransitionTime if the status of the condition doesn't change.
	if currentCond != nil && currentCond.Status == newCond.Status {
		newCond.LastTransitionTime = currentCond.LastTransitionTime
	}
	removeConditionTrial(trial, conditionType)
	trial.Status.Conditions = append(trial.Status.Conditions, newCond)
}

func removeConditionTrial(trial *morphlingv1alpha1.Trial, condType morphlingv1alpha1.TrialConditionType) {
	var newConditions []morphlingv1alpha1.TrialCondition
	for _, c := range trial.Status.Conditions {

		if c.Type == condType {
			continue
		}

		newConditions = append(newConditions, c)
	}
	trial.Status.Conditions = newConditions
}

func MarkTrialStatusCreatedTrial(trial *morphlingv1alpha1.Trial, message string) {
	SetConditionTrial(trial, morphlingv1alpha1.TrialCreated, v1.ConditionTrue, message)
}

func MarkTrialStatusSucceeded(trial *morphlingv1alpha1.Trial, status v1.ConditionStatus, message string) {
	currentCond := getConditionTrial(trial, morphlingv1alpha1.TrialRunning)
	if currentCond != nil {
		SetConditionTrial(trial, morphlingv1alpha1.TrialRunning, v1.ConditionFalse, currentCond.Message)
	}
	SetConditionTrial(trial, morphlingv1alpha1.TrialSucceeded, status, message)

}

func MarkTrialStatusFailed(trial *morphlingv1alpha1.Trial, message string) {
	currentCond := getConditionTrial(trial, morphlingv1alpha1.TrialRunning)
	if currentCond != nil {
		SetConditionTrial(trial, morphlingv1alpha1.TrialRunning, v1.ConditionFalse, currentCond.Message)
	}
	SetConditionTrial(trial, morphlingv1alpha1.TrialFailed, v1.ConditionTrue, message)
}

func MarkTrialStatusRunning(trial *morphlingv1alpha1.Trial, message string) {
	SetConditionTrial(trial, morphlingv1alpha1.TrialRunning, v1.ConditionTrue, message)
}

func GetLastConditionType(trial *morphlingv1alpha1.Trial) (morphlingv1alpha1.TrialConditionType, error) {
	if len(trial.Status.Conditions) > 0 {
		return trial.Status.Conditions[len(trial.Status.Conditions)-1].Type, nil
	}
	return "", errors.New("Trial doesn't have any condition")
}

func IsJobSucceeded(jobCondition []batchv1.JobCondition) bool {
	for _, condition := range jobCondition {
		if condition.Type == batchv1.JobComplete {
			return true
		}
	}
	return false
}

func IsServiceDeplomentReady(podConditions []appsv1.DeploymentCondition) bool {
	for _, condition := range podConditions {
		if condition.Type == appsv1.DeploymentAvailable && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func IsServiceDeplomentFail(podConditions []appsv1.DeploymentCondition) bool {
	for _, condition := range podConditions {
		if condition.Type == appsv1.DeploymentReplicaFailure && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func IsCompletedTrial(trial *morphlingv1alpha1.Trial) bool {
	return IsSucceededTrial(trial) || IsFailedTrial(trial)
}

func IsSucceededTrial(trial *morphlingv1alpha1.Trial) bool {
	return hasConditionTrial(trial, morphlingv1alpha1.TrialSucceeded)
}

func IsFailedTrial(trial *morphlingv1alpha1.Trial) bool {
	return hasConditionTrial(trial, morphlingv1alpha1.TrialFailed)
}

func IsRunningTrial(trial *morphlingv1alpha1.Trial) bool {
	return hasConditionTrial(trial, morphlingv1alpha1.TrialRunning)
}

func IsKilledTrial(trial *morphlingv1alpha1.Trial) bool {
	return hasConditionTrial(trial, morphlingv1alpha1.TrialKilled)
}

// Patch Job

// Sampling

func IsSucceededSampling(sampling *morphlingv1alpha1.Sampling) bool {
	return hasConditionSampling(sampling, morphlingv1alpha1.SamplingSucceeded)
}

func IsCreatedSampling(sampling *morphlingv1alpha1.Sampling) bool {
	return hasConditionSampling(sampling, morphlingv1alpha1.SamplingCreated)
}

func IsRunningSampling(sampling *morphlingv1alpha1.Sampling) bool {
	return hasConditionSampling(sampling, morphlingv1alpha1.SamplingRunning)
}

func MarkSamplingStatusFailed(sampling *morphlingv1alpha1.Sampling, reason, message string) {
	currentCond := getConditionSampling(sampling, morphlingv1alpha1.SamplingRunning)
	if currentCond != nil {
		SetConditionSampling(sampling, morphlingv1alpha1.SamplingRunning, v1.ConditionFalse, currentCond.Message)
	}
	SetConditionSampling(sampling, morphlingv1alpha1.SamplingFailed, v1.ConditionTrue, message)
}

func MarkSamplingStatusSucceeded(sampling *morphlingv1alpha1.Sampling, message string) {
	currentCond := getConditionSampling(sampling, morphlingv1alpha1.SamplingRunning)
	if currentCond != nil {
		SetConditionSampling(sampling, morphlingv1alpha1.SamplingRunning, v1.ConditionFalse, currentCond.Message)
	}
	SetConditionSampling(sampling, morphlingv1alpha1.SamplingSucceeded, v1.ConditionTrue, message)
}

func MarkSamplingStatusDeploymentReady(sampling *morphlingv1alpha1.Sampling, status v1.ConditionStatus, reason, message string) {
	SetConditionSampling(sampling, morphlingv1alpha1.SamplingDeploymentReady, status, message)
}

func MarkSamplingStatusRunning(sampling *morphlingv1alpha1.Sampling, reason, message string) {
	//removeCondition(SamplingRestarting)
	SetConditionSampling(sampling, morphlingv1alpha1.SamplingRunning, v1.ConditionTrue, message)
}

func MarkSamplingStatusCreated(sampling *morphlingv1alpha1.Sampling, message string) {
	//removeCondition(SamplingRestarting)
	SetConditionSampling(sampling, morphlingv1alpha1.SamplingCreated, v1.ConditionTrue, message)
}

func SetConditionSampling(sampling *morphlingv1alpha1.Sampling, conditionType morphlingv1alpha1.SamplingConditionType, status v1.ConditionStatus, message string) {
	newCond := newConditionSampling(conditionType, status, message)
	currentCond := getConditionSampling(sampling, conditionType)
	// Do nothing if condition doesn't change
	if currentCond != nil && currentCond.Status == newCond.Status {
		return
	}

	// Do not update lastTransitionTime if the status of the condition doesn't change.
	if currentCond != nil && currentCond.Status == newCond.Status {
		return
	}
	removeConditionSampling(sampling, conditionType)
	sampling.Status.Conditions = append(sampling.Status.Conditions, newCond)
}

func removeConditionSampling(sampling *morphlingv1alpha1.Sampling, condType morphlingv1alpha1.SamplingConditionType) {
	var newConditions []morphlingv1alpha1.SamplingCondition
	for _, c := range sampling.Status.Conditions {

		if c.Type == condType {
			continue
		}

		newConditions = append(newConditions, c)
	}
	sampling.Status.Conditions = newConditions
}

func newConditionSampling(conditionType morphlingv1alpha1.SamplingConditionType, status v1.ConditionStatus, message string) morphlingv1alpha1.SamplingCondition {
	return morphlingv1alpha1.SamplingCondition{
		Type:           conditionType,
		Status:         status,
		LastUpdateTime: metav1.Now(),
		Message:        message,
	}
}

func getConditionSampling(sampling *morphlingv1alpha1.Sampling, condType morphlingv1alpha1.SamplingConditionType) *morphlingv1alpha1.SamplingCondition {

	if sampling.Status.Conditions != nil {
		for _, condition := range sampling.Status.Conditions {
			if condition.Type == condType {
				return &condition
			}
		}
	}
	return nil
}

// Sampling service
//type General struct {
//	scheme *runtime.Scheme
//	client.Client
//}

//func (g *General) DesiredService(s *Sampling) (*corev1.Service, error) {
//	ports := []corev1.ServicePort{
//		{
//			Name: consts.DefaultSamplingPortName,
//			Port: consts.DefaultSamplingPort,
//		},
//	}
//
//	service := &corev1.Service{
//		ObjectMeta: metav1.ObjectMeta{
//			Name:      s.Name + "-" + string(s.Spec.Algorithm.AlgorithmName),//util.GetAlgorithmServiceName(s),
//			Namespace: s.Namespace,
//		},
//		Spec: corev1.ServiceSpec{
//			Selector: SamplingLabels(s),
//			Ports:    ports,
//			Type:     corev1.ServiceTypeClusterIP,
//		},
//	}
//
//	// Add owner reference to the service so that it could be GC after the sampling is deleted
//	if err := controllerutil.SetControllerReference(s, service, g.scheme); err != nil {
//		return nil, err
//	}
//
//	return service, nil
//}

//func SamplingLabels(instance *morphlingv1alpha1.Sampling) map[string]string {
//	res := make(map[string]string)
//	for k, v := range instance.Labels {
//		res[k] = v
//	}
//	res[consts.LabelDeploymentName] = instance.Name + "-" + string(instance.Spec.Algorithm.AlgorithmName)
//	res[consts.LabelExperimentName] = instance.Name
//	res[consts.LabelSamplingName] = instance.Name
//
//	return res
//}
