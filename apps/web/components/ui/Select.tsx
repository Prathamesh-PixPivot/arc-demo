import { SelectHTMLAttributes, forwardRef } from 'react';
import { cn } from '@/lib/utils';

export interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
    label?: string;
    error?: string;
    helperText?: string;
    options: { value: string; label: string }[];
}

const Select = forwardRef<HTMLSelectElement, SelectProps>(
    ({ className, label, error, helperText, id, options, ...props }, ref) => {
        const selectId = id || label?.toLowerCase().replace(/\s+/g, '-');

        return (
            <div className="w-full">
                {label && (
                    <label
                        htmlFor={selectId}
                        className="block text-sm font-medium text-gray-700 mb-1.5"
                    >
                        {label}
                        {props.required && <span className="text-error ml-1" aria-label="required">*</span>}
                    </label>
                )}
                <select
                    ref={ref}
                    id={selectId}
                    className={cn(
                        'h-10 w-full rounded-md border border-gray-200 px-3 py-2 text-sm focus-ring focus:border-purple transition-colors bg-white',
                        error && 'border-error focus:ring-error',
                        className
                    )}
                    aria-invalid={error ? 'true' : 'false'}
                    aria-describedby={error ? `${selectId}-error` : helperText ? `${selectId}-helper` : undefined}
                    {...props}
                >
                    <option value="">Select an option</option>
                    {options.map((option) => (
                        <option key={option.value} value={option.value}>
                            {option.label}
                        </option>
                    ))}
                </select>
                {error && (
                    <p
                        id={`${selectId}-error`}
                        className="mt-1.5 text-sm text-error"
                        role="alert"
                    >
                        {error}
                    </p>
                )}
                {helperText && !error && (
                    <p
                        id={`${selectId}-helper`}
                        className="mt-1.5 text-sm text-gray-500"
                    >
                        {helperText}
                    </p>
                )}
            </div>
        );
    }
);

Select.displayName = 'Select';

export { Select };
