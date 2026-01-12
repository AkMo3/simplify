import { clsx, type ClassValue } from 'clsx'
import { twMerge } from 'tailwind-merge'

/**
 * Merge Tailwind CSS classes with clsx
 * This is the standard pattern used by shadcn/ui and Aceternity
 */
export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
