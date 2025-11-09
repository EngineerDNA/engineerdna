import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/client';
import { WelcomeStep } from './onboarding/welcome-step';
import { RoleSelectionStep } from './onboarding/role-selection-step';
import { SourceConfigStep } from './onboarding/source-config-step';
import { FirstSyncStep } from './onboarding/first-sync-step';
import type { UserRole } from '../api/types';

interface OnboardingModalProps {
  onComplete: () => void;
}

type OnboardingStep = 'welcome' | 'role' | 'sources' | 'sync' | 'complete';

export function OnboardingModal({ onComplete }: OnboardingModalProps) {
  const [currentStep, setCurrentStep] = useState<OnboardingStep>('welcome');
  const [selectedRole, setSelectedRole] = useState<UserRole | null>(null);
  const queryClient = useQueryClient();

  const completeMutation = useMutation({
    mutationFn: api.completeOnboarding,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['onboarding-status'] });
      onComplete();
    },
  });

  const handleSkip = () => {
    completeMutation.mutate();
  };

  const handleComplete = () => {
    completeMutation.mutate();
  };

  const handleRoleSelect = (role: UserRole) => {
    setSelectedRole(role);
  };

  const steps: OnboardingStep[] = ['welcome', 'role', 'sources', 'sync', 'complete'];
  const currentStepIndex = steps.indexOf(currentStep);

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-3xl w-full max-h-[90vh] flex flex-col">
        <div className="p-6 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center justify-between mb-4">
            <h1 className="text-xl font-bold text-gray-900 dark:text-gray-100">
              Setup EngineerDNA
            </h1>
            <button
              onClick={handleSkip}
              className="text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200 transition-colors"
              aria-label="Close"
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M6 18L18 6M6 6l12 12"
                />
              </svg>
            </button>
          </div>

          <div className="flex items-center gap-2">
            {steps.map((step, index) => (
              <div key={step} className="flex items-center flex-1">
                <div
                  className={`flex-1 h-2 rounded-full transition-colors ${
                    index <= currentStepIndex
                      ? 'bg-blue-600 dark:bg-blue-500'
                      : 'bg-gray-200 dark:bg-gray-700'
                  }`}
                />
              </div>
            ))}
          </div>
        </div>

        <div className="flex-1 overflow-y-auto p-6">
          {currentStep === 'welcome' && (
            <WelcomeStep onNext={() => setCurrentStep('role')} onSkip={handleSkip} />
          )}
          {currentStep === 'role' && (
            <RoleSelectionStep
              selectedRole={selectedRole}
              onRoleSelect={handleRoleSelect}
              onNext={() => setCurrentStep('sources')}
              onBack={() => setCurrentStep('welcome')}
            />
          )}
          {currentStep === 'sources' && (
            <SourceConfigStep
              onNext={() => setCurrentStep('sync')}
              onSkip={handleSkip}
              onBack={() => setCurrentStep('role')}
            />
          )}
          {currentStep === 'sync' && (
            <FirstSyncStep
              selectedRole={selectedRole}
              onComplete={() => setCurrentStep('complete')}
              onBack={() => setCurrentStep('sources')}
            />
          )}
          {currentStep === 'complete' && (
            <div className="space-y-6 text-center py-8">
              <div className="flex justify-center">
                <svg
                  className="w-20 h-20 text-green-600 dark:text-green-400"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"
                  />
                </svg>
              </div>
              <div>
                <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-2">
                  Setup Complete!
                </h2>
                <p className="text-gray-600 dark:text-gray-400">
                  Your EngineerDNA workspace is ready. Start exploring your dashboards and insights.
                </p>
              </div>
              <div className="flex justify-center pt-4">
                <button
                  onClick={handleComplete}
                  className="px-8 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 dark:bg-blue-700 dark:hover:bg-blue-800 transition-colors focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 dark:focus:ring-offset-gray-900 font-medium text-lg"
                >
                  Go to Dashboards
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
