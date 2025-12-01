'use client';

import { useState, useEffect } from 'react';
import { cn } from '@/lib/utils';

interface PasswordStrengthProps {
    password: string;
    className?: string;
}

export function PasswordStrength({ password, className }: PasswordStrengthProps) {
    const [strength, setStrength] = useState<'weak' | 'medium' | 'strong' | null>(null);
    const [score, setScore] = useState(0);

    useEffect(() => {
        if (!password) {
            setStrength(null);
            setScore(0);
            return;
        }

        let points = 0;

        // Length
        if (password.length >= 8) points += 1;
        if (password.length >= 12) points += 1;
        if (password.length >= 16) points += 1;

        // Contains lowercase
        if (/[a-z]/.test(password)) points += 1;

        // Contains uppercase
        if (/[A-Z]/.test(password)) points += 1;

        // Contains numbers
        if (/\d/.test(password)) points += 1;

        // Contains special characters
        if (/[^a-zA-Z0-9]/.test(password)) points += 1;

        setScore(points);

        if (points <= 3) {
            setStrength('weak');
        } else if (points <= 5) {
            setStrength('medium');
        } else {
            setStrength('strong');
        }
    }, [password]);

    if (!password) return null;

    const strengthColors = {
        weak: 'bg-error',
        medium: 'bg-warning',
        strong: 'bg-success',
    };

    const strengthLabels = {
        weak: 'Weak',
        medium: 'Medium',
        strong: 'Strong',
    };

    return (
        <div className={cn('space-y-2', className)}>
            <div className="flex gap-1">
                {[1, 2, 3, 4].map((level) => (
                    <div
                        key={level}
                        className={cn(
                            'h-1 flex-1 rounded-full transition-colors',
                            score >= level * 1.75
                                ? strength && strengthColors[strength]
                                : 'bg-gray-200'
                        )}
                    />
                ))}
            </div>
            {strength && (
                <p className="text-xs text-gray-600">
                    Password strength: <span className={cn('font-medium', {
                        'text-error': strength === 'weak',
                        'text-warning': strength === 'medium',
                        'text-success': strength === 'strong',
                    })}>
                        {strengthLabels[strength]}
                    </span>
                </p>
            )}
            <ul className="text-xs text-gray-600 space-y-1">
                <li className={password.length >= 8 ? 'text-success' : ''}>
                    ✓ At least 8 characters
                </li>
                <li className={/[a-z]/.test(password) && /[A-Z]/.test(password) ? 'text-success' : ''}>
                    ✓ Uppercase and lowercase letters
                </li>
                <li className={/\d/.test(password) ? 'text-success' : ''}>
                    ✓ At least one number
                </li>
                <li className={/[^a-zA-Z0-9]/.test(password) ? 'text-success' : ''}>
                    ✓ At least one special character
                </li>
            </ul>
        </div>
    );
}
