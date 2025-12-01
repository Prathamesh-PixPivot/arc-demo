import { HTMLAttributes, forwardRef } from 'react';
import { cn } from '@/lib/utils';

export interface BadgeProps extends HTMLAttributes<HTMLDivElement> {
    variant?: 'success' | 'warning' | 'error' | 'info' | 'gray' | 'purple';
    size?: 'sm' | 'md';
}

const Badge = forwardRef<HTMLDivElement, BadgeProps>(
    ({ className, variant = 'gray', size = 'md', children, ...props }, ref) => {
        const variantStyles = {
            success: 'badge-success',
            warning: 'badge-warning',
            error: 'badge-error',
            info: 'badge-info',
            gray: 'badge-gray',
            purple: 'bg-purple-light text-purple-dark',
        };

        const sizeStyles = {
            sm: 'text-xs px-2 py-0.5',
            md: 'text-sm px-2.5 py-0.5',
        };

        return (
            <div
                ref={ref}
                className={cn('badge', variantStyles[variant], sizeStyles[size], className)}
                {...props}
            >
                {children}
            </div>
        );
    }
);

Badge.displayName = 'Badge';

export { Badge };
