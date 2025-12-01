'use client';

import { cn } from '@/lib/utils';

interface Step {
    id: number;
    name: string;
    description?: string;
}

interface StepIndicatorProps {
    steps: Step[];
    currentStep: number;
    className?: string;
}

export function StepIndicator({ steps, currentStep, className }: StepIndicatorProps) {
    return (
        <nav aria-label="Progress" className={cn('', className)}>
            <ol role="list" className="flex items-center justify-between">
                {steps.map((step, stepIdx) => (
                    <li
                        key={step.id}
                        className={cn(
                            'relative',
                            stepIdx !== steps.length - 1 ? 'pr-8 sm:pr-20 flex-1' : ''
                        )}
                    >
                        {/* Connector line */}
                        {stepIdx !== steps.length - 1 && (
                            <div
                                className="absolute inset-0 flex items-center top-4"
                                aria-hidden="true"
                            >
                                <div
                                    className={cn(
                                        'h-0.5 w-full transition-colors',
                                        currentStep > step.id ? 'bg-purple' : 'bg-gray-200'
                                    )}
                                />
                            </div>
                        )}

                        {/* Step circle */}
                        <div className="relative flex flex-col items-center group">
                            <span
                                className={cn(
                                    'relative z-10 w-8 h-8 flex items-center justify-center rounded-full text-sm font-medium transition-colors',
                                    currentStep > step.id
                                        ? 'bg-purple text-white'
                                        : currentStep === step.id
                                            ? 'bg-purple text-white ring-4 ring-purple-light'
                                            : 'bg-white border-2 border-gray-300 text-gray-500'
                                )}
                            >
                                {currentStep > step.id ? (
                                    <svg
                                        className="w-5 h-5"
                                        xmlns="http://www.w3.org/2000/svg"
                                        viewBox="0 0 20 20"
                                        fill="currentColor"
                                    >
                                        <path
                                            fillRule="evenodd"
                                            d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
                                            clipRule="evenodd"
                                        />
                                    </svg>
                                ) : (
                                    <span>{step.id}</span>
                                )}
                            </span>
                            <span
                                className={cn(
                                    'mt-2 text-xs font-medium text-center',
                                    currentStep >= step.id ? 'text-purple' : 'text-gray-500'
                                )}
                            >
                                {step.name}
                            </span>
                            {step.description && (
                                <span className="mt-0.5 text-xs text-gray-500 text-center hidden sm:block">
                                    {step.description}
                                </span>
                            )}
                        </div>
                    </li>
                ))}
            </ol>
        </nav>
    );
}
