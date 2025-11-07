import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../api/client';
import { WelcomeStep } from './onboarding/welcome-step';
import { TeamImportStep } from './onboarding/team-import-step';
import { SourceConfigStep } from './onboarding/source-config-step';
import { FirstSyncStep } from './onboarding/first-sync-step';

interface OnboardingModalProps {
  onComplete: () => void;
}

type OnboardingStep = 'welcome' | 'team' | 'sources' | 'sync';

export function OnboardingModal({ onComplete }: OnboardingModalProps) {
  const [currentStep, setCurrentStep] = useState<OnboardingStep>('welcome');
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

  const steps: OnboardingStep[] = ['welcome', 'team', 'sources', 'sync'];
  const currentStepIndex = steps.indexOf(currentStep);

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-3xl w-full max-h-[90vh] overflow-hidden flex flex-col">
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
            <WelcomeStep onNext={() => setCurrentStep('team')} onSkip={handleSkip} />
          )}
          {currentStep === 'team' && (
            <TeamImportStep
              onNext={() => setCurrentStep('sources')}
              onSkip={handleSkip}
              onBack={() => setCurrentStep('welcome')}
            />
          )}
          {currentStep === 'sources' && (
            <SourceConfigStep
              onNext={() => setCurrentStep('sync')}
              onSkip={handleSkip}
              onBack={() => setCurrentStep('team')}
            />
          )}
          {currentStep === 'sync' && (
            <FirstSyncStep onComplete={handleComplete} onBack={() => setCurrentStep('sources')} />
          )}
        </div>
      </div>
    </div>
  );
}
