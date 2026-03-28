'use client';

import { cn } from '@/lib/utils';
import { FormField } from './form-field';
import { ChevronDown } from 'lucide-react';

// ── Types ────────────────────────────────────────────────
interface SelectOption {
  value: string;
  label: string;
}

interface SelectFieldProps {
  label: string;
  name: string;
  options: SelectOption[];
  value: string;
  onChange: (value: string) => void;
  error?: string;
  placeholder?: string;
  required?: boolean;
  disabled?: boolean;
  className?: string;
}

// ── Component ────────────────────────────────────────────
export function SelectField({
  label,
  name,
  options,
  value,
  onChange,
  error,
  placeholder = 'Select an option',
  required,
  disabled,
  className,
}: SelectFieldProps) {
  return (
    <FormField
      label={label}
      name={name}
      error={error}
      required={required}
      className={className}
    >
      <div className="relative">
        <select
          id={name}
          name={name}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          disabled={disabled}
          className={cn(
            'input appearance-none pr-10',
            error && 'border-red-300 focus:border-red-500 focus:ring-red-500',
            disabled && 'cursor-not-allowed bg-gray-50 text-gray-500',
            !value && 'text-gray-400',
          )}
        >
          <option value="" disabled>
            {placeholder}
          </option>
          {options.map((opt) => (
            <option key={opt.value} value={opt.value}>
              {opt.label}
            </option>
          ))}
        </select>
        <ChevronDown className="pointer-events-none absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400" />
      </div>
    </FormField>
  );
}
