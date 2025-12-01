import { TextareaHTMLAttributes, forwardRef } from 'react';
import { cn } from '@/lib/utils';

export interface TextareaProps extends TextareaHTMLAttributes<HTMLTextAreaElement> {
    label?: string;
    error?: string;
    helperText?: string;
}

const Textarea = forwardRef<HTMLTextAreaElement, TextareaProps>(
    ({ className, label, error, helperText, id, ...props }, ref) => {
        const textareaId = id || label?.toLowerCase().replace(/\s+/g, '-');

        return (
            <div className="w-full">
                {label && (
                    <label
                        htmlFor={textareaId}
                        className="block text-sm font-medium text-gray-700 mb-1.5"
                    >
                        {label}
                        {props.required && <span className="text-error ml-1" aria-label="required">*</span>}
                    </label>
                )}
                <textarea
                    ref={ref}
                    id={textareaId}
                    className={cn(
                        'flex min-h-[80px] w-full rounded-md border border-gray-200 px-3 py-2 text-sm focus-ring focus:border-purple transition-colors disabled:cursor-not-allowed disabled:opacity-50',
                        error && 'border-error focus:ring-error',
                        className
                    )}
                    aria-invalid={error ? 'true' : 'false'}
                    aria-describedby={error ? `${textareaId}-error` : helperText ? `${textareaId}-helper` : undefined}
                    {...props}
                />
                {error && (
                    <p
                        id={`${textareaId}-error`}
                        className="mt-1.5 text-sm text-error"
                        role="alert"
                    >
                        {error}
                    </p>
                )}
                {helperText && !error && (
                    <p
                        id={`${textareaId}-helper`}
                        className="mt-1.5 text-sm text-gray-500"
                    >
                        {helperText}
                    </p>
                )}
            </div>
        );
    }
);

Textarea.displayName = 'Textarea';

export { Textarea };
