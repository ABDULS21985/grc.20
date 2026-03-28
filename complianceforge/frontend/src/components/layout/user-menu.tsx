'use client';

import * as DropdownMenu from '@radix-ui/react-dropdown-menu';
import {
  User as UserIcon,
  Settings,
  ClipboardList,
  LogOut,
} from 'lucide-react';
import Link from 'next/link';

// ── Types ────────────────────────────────────────────────
interface UserMenuUser {
  first_name: string;
  last_name: string;
  email: string;
}

interface UserMenuProps {
  user: UserMenuUser;
}

// ── Component ────────────────────────────────────────────
export function UserMenu({ user }: UserMenuProps) {
  const initials = `${user.first_name.charAt(0)}${user.last_name.charAt(0)}`.toUpperCase();

  return (
    <DropdownMenu.Root>
      <DropdownMenu.Trigger asChild>
        <button
          className="flex h-9 w-9 items-center justify-center rounded-full bg-indigo-100 text-xs font-semibold text-indigo-700 ring-offset-2 hover:ring-2 hover:ring-indigo-300 focus:outline-none focus:ring-2 focus:ring-indigo-500 transition-all"
          aria-label="User menu"
        >
          {initials}
        </button>
      </DropdownMenu.Trigger>

      <DropdownMenu.Portal>
        <DropdownMenu.Content
          align="end"
          sideOffset={8}
          className="z-50 min-w-[220px] rounded-xl border border-gray-200 bg-white p-1.5 shadow-lg animate-in fade-in slide-in-from-top-2"
        >
          {/* User Info Header */}
          <div className="px-3 py-2.5 border-b border-gray-100 mb-1">
            <p className="text-sm font-medium text-gray-900">
              {user.first_name} {user.last_name}
            </p>
            <p className="text-xs text-gray-500 truncate">{user.email}</p>
          </div>

          <DropdownMenu.Item asChild>
            <Link
              href="/settings/profile"
              className="flex items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-gray-700 outline-none hover:bg-gray-50 focus:bg-gray-50 cursor-pointer"
            >
              <UserIcon className="h-4 w-4 text-gray-400" />
              Profile
            </Link>
          </DropdownMenu.Item>

          <DropdownMenu.Item asChild>
            <Link
              href="/settings/organisation"
              className="flex items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-gray-700 outline-none hover:bg-gray-50 focus:bg-gray-50 cursor-pointer"
            >
              <Settings className="h-4 w-4 text-gray-400" />
              Organisation Settings
            </Link>
          </DropdownMenu.Item>

          <DropdownMenu.Item asChild>
            <Link
              href="/settings/audit-log"
              className="flex items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-gray-700 outline-none hover:bg-gray-50 focus:bg-gray-50 cursor-pointer"
            >
              <ClipboardList className="h-4 w-4 text-gray-400" />
              Audit Log
            </Link>
          </DropdownMenu.Item>

          <DropdownMenu.Separator className="my-1 h-px bg-gray-100" />

          <DropdownMenu.Item asChild>
            <Link
              href="/auth/logout"
              className="flex items-center gap-2.5 rounded-lg px-3 py-2 text-sm text-red-600 outline-none hover:bg-red-50 focus:bg-red-50 cursor-pointer"
            >
              <LogOut className="h-4 w-4 text-red-400" />
              Logout
            </Link>
          </DropdownMenu.Item>
        </DropdownMenu.Content>
      </DropdownMenu.Portal>
    </DropdownMenu.Root>
  );
}
