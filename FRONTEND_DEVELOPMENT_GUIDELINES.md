# Kainuguru Frontend Development Guidelines

## Executive Summary

This document provides comprehensive guidelines for building the Kainuguru frontend applications using a monorepo architecture with React Native (Expo) for mobile and Next.js for web. The architecture prioritizes code reuse while maintaining platform-specific optimizations, with all business logic remaining in the Go backend.

**Key Principles:**
- Go backend handles ALL business logic and mutations
- Next.js Server Components for data fetching only (no Server Actions for business logic)
- Maximum code reuse through shared packages
- Platform-specific UI libraries (shadcn/ui for web, Tamagui/NativeWind for mobile)
- Type-safe end-to-end with TypeScript and GraphQL codegen

---

## 1. Technology Stack

### Core Technologies

| Category | Web | Mobile | Shared |
|----------|-----|--------|--------|
| **Framework** | Next.js 15 (App Router) | Expo SDK 54+ | React 19 |
| **Language** | TypeScript 5.6+ | TypeScript 5.6+ | TypeScript 5.6+ |
| **UI Components** | shadcn/ui + Radix UI | Tamagui or NativeWind | Design tokens |
| **Styling** | Tailwind CSS 3.4+ | Tamagui styles or NativeWind | Color/spacing tokens |
| **State Management** | Zustand | Zustand | Zustand |
| **API Client** | urql or Apollo Client | urql or Apollo Client | GraphQL Codegen |
| **Navigation** | Next.js App Router | Expo Router | Solito (optional) |
| **Forms** | React Hook Form + Zod | React Hook Form + Zod | Shared validation |
| **Auth Storage** | HTTP-only cookies | Expo SecureStore | Auth utilities |
| **Build Tool** | Turbopack | Metro | Turborepo |
| **Package Manager** | pnpm | pnpm | pnpm |
| **Testing** | Vitest + Playwright | Jest + Detox | Shared test utils |
| **Deployment** | Vercel | EAS Build/Submit | GitHub Actions |

### Library Versions (Lock these in package.json)

```json
{
  "dependencies": {
    "react": "^19.0.0",
    "react-native": "0.76.0",
    "expo": "~54.0.0",
    "next": "15.0.0",
    "@radix-ui/react-*": "^1.1.0",
    "tailwindcss": "^3.4.0",
    "zustand": "^5.0.0",
    "urql": "^4.1.0",
    "@tamagui/core": "^1.110.0",
    "nativewind": "^4.1.0",
    "react-hook-form": "^7.54.0",
    "zod": "^3.24.0"
  }
}
```

---

## 2. Monorepo Structure

```
kainuguru-frontend/
├── apps/
│   ├── web/                          # Next.js application
│   │   ├── app/                      # App Router pages
│   │   │   ├── (auth)/              # Auth group routes
│   │   │   │   ├── login/
│   │   │   │   └── register/
│   │   │   ├── (dashboard)/         # Dashboard group
│   │   │   │   ├── layout.tsx       # Dashboard layout
│   │   │   │   ├── page.tsx         # Dashboard home
│   │   │   │   └── lists/
│   │   │   │       ├── page.tsx     # Lists page (RSC)
│   │   │   │       ├── [id]/
│   │   │   │       │   └── page.tsx # List detail (RSC)
│   │   │   │       └── new/
│   │   │   ├── api/                 # Route handlers (proxies only)
│   │   │   │   ├── auth/
│   │   │   │   │   ├── login/route.ts
│   │   │   │   │   └── refresh/route.ts
│   │   │   │   └── proxy/           # GraphQL proxy
│   │   │   │       └── [...path]/route.ts
│   │   │   ├── layout.tsx           # Root layout
│   │   │   ├── error.tsx
│   │   │   ├── loading.tsx
│   │   │   └── providers.tsx        # Client providers
│   │   ├── components/
│   │   │   ├── ui/                  # shadcn/ui components
│   │   │   │   ├── button.tsx
│   │   │   │   ├── card.tsx
│   │   │   │   ├── dialog.tsx
│   │   │   │   └── ...
│   │   │   └── features/            # Feature components
│   │   │       ├── shopping-list/
│   │   │       └── auth/
│   │   ├── lib/
│   │   │   ├── api-client.ts        # Web-specific API setup
│   │   │   └── utils.ts             # Web utilities
│   │   ├── styles/
│   │   │   └── globals.css          # Tailwind imports
│   │   ├── public/
│   │   ├── next.config.js
│   │   ├── tailwind.config.ts
│   │   ├── components.json          # shadcn/ui config
│   │   └── package.json
│   │
│   └── mobile/                       # Expo application
│       ├── app/                      # Expo Router screens
│       │   ├── (auth)/
│       │   │   ├── login.tsx
│       │   │   └── register.tsx
│       │   ├── (tabs)/
│       │   │   ├── _layout.tsx      # Tab layout
│       │   │   ├── index.tsx        # Home tab
│       │   │   ├── lists.tsx        # Lists tab
│       │   │   └── profile.tsx      # Profile tab
│       │   ├── lists/
│       │   │   └── [id].tsx         # List detail
│       │   ├── _layout.tsx          # Root layout
│       │   └── +not-found.tsx
│       ├── components/
│       │   ├── ui/                  # Native UI components
│       │   │   ├── Button.tsx
│       │   │   ├── Card.tsx
│       │   │   └── ...
│       │   └── features/
│       │       ├── barcode-scanner/
│       │       └── shopping-list/
│       ├── lib/
│       │   ├── api-client.ts        # Mobile-specific API
│       │   └── secure-store.ts      # Auth token storage
│       ├── assets/
│       ├── app.json
│       ├── eas.json
│       ├── metro.config.js
│       ├── tamagui.config.ts        # OR nativewind.config.js
│       └── package.json
│
├── packages/
│   ├── api-client/                  # Generated GraphQL client
│   │   ├── src/
│   │   │   ├── generated/           # GraphQL Codegen output
│   │   │   │   ├── types.ts        # Generated types
│   │   │   │   ├── operations.ts   # Generated operations
│   │   │   │   └── index.ts
│   │   │   ├── client.ts            # urql client setup
│   │   │   ├── queries/             # .graphql files
│   │   │   │   ├── shopping-lists.graphql
│   │   │   │   └── auth.graphql
│   │   │   ├── mutations/
│   │   │   │   ├── shopping-lists.graphql
│   │   │   │   └── auth.graphql
│   │   │   ├── subscriptions/
│   │   │   └── hooks/               # Custom React hooks
│   │   │       ├── useShoppingList.ts
│   │   │       └── useAuth.ts
│   │   ├── codegen.yml
│   │   └── package.json
│   │
│   ├── tokens/                      # Design tokens
│   │   ├── src/
│   │   │   ├── colors.ts            # Brand colors
│   │   │   ├── spacing.ts          # Spacing scale
│   │   │   ├── typography.ts       # Font sizes/weights
│   │   │   ├── radii.ts            # Border radius
│   │   │   └── index.ts
│   │   └── package.json
│   │
│   ├── utils/                       # Shared utilities
│   │   ├── src/
│   │   │   ├── validation/          # Zod schemas
│   │   │   │   ├── shopping-list.ts
│   │   │   │   └── auth.ts
│   │   │   ├── formatting/          # Data formatters
│   │   │   │   ├── date.ts
│   │   │   │   └── currency.ts
│   │   │   ├── constants/           # App constants
│   │   │   └── types/               # Shared TypeScript types
│   │   │       └── index.ts
│   │   └── package.json
│   │
│   ├── stores/                      # Zustand stores
│   │   ├── src/
│   │   │   ├── auth.store.ts
│   │   │   ├── shopping-list.store.ts
│   │   │   └── ui.store.ts
│   │   └── package.json
│   │
│   └── config/                      # Shared configs
│       ├── eslint/
│       │   └── index.js
│       ├── typescript/
│       │   └── base.json
│       └── prettier/
│           └── index.js
│
├── .github/
│   └── workflows/
│       ├── ci.yml
│       ├── deploy-web.yml
│       └── build-mobile.yml
│
├── turbo.json                       # Turborepo config
├── package.json                     # Root package.json
├── pnpm-workspace.yaml             # pnpm workspace config
├── .env.example
├── .gitignore
├── README.md
└── FRONTEND_DEVELOPMENT_GUIDELINES.md  # This document
```

---

## 3. Development Patterns

### 3.1 API Integration (Go Backend)

**CRITICAL: All business logic stays in Go. Next.js is ONLY for rendering and proxying.**

#### Web: Server Components for Data Fetching

```typescript
// apps/web/app/(dashboard)/lists/page.tsx
import { getClient } from '@/lib/api-client'
import { GetShoppingListsDocument } from '@kainuguru/api-client'

// This runs on the server - data fetching only
export default async function ListsPage() {
  const client = getClient() // Server-side GraphQL client
  const { data } = await client.query(GetShoppingListsDocument, {}).toPromise()

  return (
    <div className="container mx-auto p-4">
      <h1 className="text-3xl font-bold mb-6">Shopping Lists</h1>
      {/* Render the data */}
      <ShoppingListGrid lists={data?.shoppingLists} />
    </div>
  )
}
```

#### Web: Route Handler as Proxy (When Needed)

```typescript
// apps/web/app/api/proxy/[...path]/route.ts
import { NextRequest, NextResponse } from 'next/server'
import { cookies } from 'next/headers'

const GO_API_URL = process.env.GO_API_URL || 'http://localhost:4000'

// Proxy to handle cookies/auth for same-origin
export async function POST(
  request: NextRequest,
  { params }: { params: { path: string[] } }
) {
  const path = params.path.join('/')
  const cookieStore = cookies()
  const authToken = cookieStore.get('auth-token')

  const response = await fetch(`${GO_API_URL}/${path}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(authToken && { 'Authorization': `Bearer ${authToken.value}` }),
    },
    body: await request.text(),
  })

  const data = await response.json()
  return NextResponse.json(data, { status: response.status })
}
```

#### Mobile: Direct API Calls

```typescript
// apps/mobile/lib/api-client.ts
import { createClient } from '@kainuguru/api-client'
import * as SecureStore from 'expo-secure-store'

const GO_API_URL = process.env.EXPO_PUBLIC_API_URL

export const apiClient = createClient({
  url: GO_API_URL,
  fetchOptions: async () => {
    const token = await SecureStore.getItemAsync('auth-token')
    return {
      headers: {
        authorization: token ? `Bearer ${token}` : '',
      },
    }
  },
})
```

#### Shared: GraphQL Operations

```graphql
# packages/api-client/src/queries/shopping-lists.graphql
query GetShoppingLists($limit: Int, $offset: Int) {
  shoppingLists(limit: $limit, offset: $offset) {
    edges {
      node {
        id
        name
        description
        itemCount
        completedItemCount
      }
    }
    pageInfo {
      hasNextPage
      totalCount
    }
  }
}

mutation CreateShoppingList($input: CreateShoppingListInput!) {
  createShoppingList(input: $input) {
    id
    name
  }
}
```

### 3.2 Authentication Patterns

#### Web: HTTP-Only Cookies

```typescript
// apps/web/app/api/auth/login/route.ts
import { NextRequest, NextResponse } from 'next/server'

export async function POST(request: NextRequest) {
  const body = await request.json()

  // Call Go auth endpoint
  const response = await fetch(`${process.env.GO_API_URL}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })

  const data = await response.json()

  if (response.ok) {
    const res = NextResponse.json({ success: true, user: data.user })

    // Set HTTP-only cookie
    res.cookies.set('auth-token', data.token, {
      httpOnly: true,
      secure: process.env.NODE_ENV === 'production',
      sameSite: 'lax',
      maxAge: 60 * 60 * 24 * 7, // 7 days
    })

    return res
  }

  return NextResponse.json(
    { error: data.error },
    { status: response.status }
  )
}
```

#### Mobile: Secure Token Storage

```typescript
// apps/mobile/lib/auth.ts
import * as SecureStore from 'expo-secure-store'

export class AuthService {
  private static TOKEN_KEY = 'auth-token'
  private static REFRESH_KEY = 'refresh-token'

  static async login(email: string, password: string) {
    const response = await fetch(`${API_URL}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    })

    const data = await response.json()

    if (response.ok) {
      await SecureStore.setItemAsync(this.TOKEN_KEY, data.accessToken)
      await SecureStore.setItemAsync(this.REFRESH_KEY, data.refreshToken)
      return { success: true, user: data.user }
    }

    throw new Error(data.error)
  }

  static async refreshToken() {
    const refreshToken = await SecureStore.getItemAsync(this.REFRESH_KEY)

    if (!refreshToken) {
      throw new Error('No refresh token')
    }

    const response = await fetch(`${API_URL}/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refreshToken }),
    })

    const data = await response.json()

    if (response.ok) {
      await SecureStore.setItemAsync(this.TOKEN_KEY, data.accessToken)
      return data.accessToken
    }

    throw new Error('Failed to refresh token')
  }
}
```

### 3.3 UI Component Patterns

#### Web: shadcn/ui Components

```typescript
// apps/web/components/ui/button.tsx
import * as React from "react"
import { Slot } from "@radix-ui/react-slot"
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "@/lib/utils"

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50",
  {
    variants: {
      variant: {
        default: "bg-primary text-primary-foreground shadow hover:bg-primary/90",
        destructive: "bg-destructive text-destructive-foreground shadow-sm hover:bg-destructive/90",
        outline: "border border-input bg-background shadow-sm hover:bg-accent hover:text-accent-foreground",
        secondary: "bg-secondary text-secondary-foreground shadow-sm hover:bg-secondary/80",
        ghost: "hover:bg-accent hover:text-accent-foreground",
        link: "text-primary underline-offset-4 hover:underline",
      },
      size: {
        default: "h-9 px-4 py-2",
        sm: "h-8 rounded-md px-3 text-xs",
        lg: "h-10 rounded-md px-8",
        icon: "h-9 w-9",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  }
)

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {
  asChild?: boolean
}

const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, ...props }, ref) => {
    const Comp = asChild ? Slot : "button"
    return (
      <Comp
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        {...props}
      />
    )
  }
)
Button.displayName = "Button"

export { Button, buttonVariants }
```

#### Mobile: Tamagui Components

```typescript
// apps/mobile/components/ui/Button.tsx
import { Button as TamaguiButton, ButtonProps as TamaguiButtonProps } from '@tamagui/button'
import { styled } from '@tamagui/core'

export const Button = styled(TamaguiButton, {
  name: 'Button',

  variants: {
    variant: {
      primary: {
        backgroundColor: '$primary',
        color: '$primaryForeground',
      },
      secondary: {
        backgroundColor: '$secondary',
        color: '$secondaryForeground',
      },
      destructive: {
        backgroundColor: '$destructive',
        color: '$destructiveForeground',
      },
      outline: {
        backgroundColor: 'transparent',
        borderWidth: 1,
        borderColor: '$border',
      },
    },

    size: {
      sm: {
        paddingHorizontal: '$3',
        paddingVertical: '$2',
        fontSize: '$2',
      },
      md: {
        paddingHorizontal: '$4',
        paddingVertical: '$3',
        fontSize: '$3',
      },
      lg: {
        paddingHorizontal: '$6',
        paddingVertical: '$4',
        fontSize: '$4',
      },
    },
  } as const,

  defaultVariants: {
    variant: 'primary',
    size: 'md',
  },
})
```

#### Mobile: NativeWind Alternative

```typescript
// apps/mobile/components/ui/Button.tsx
import { Pressable, Text, PressableProps } from 'react-native'
import { cva, type VariantProps } from 'class-variance-authority'

const buttonVariants = cva(
  'flex-row items-center justify-center rounded-md',
  {
    variants: {
      variant: {
        primary: 'bg-blue-600',
        secondary: 'bg-gray-200',
        destructive: 'bg-red-600',
        outline: 'border border-gray-300',
      },
      size: {
        sm: 'px-3 py-1.5',
        md: 'px-4 py-2',
        lg: 'px-6 py-3',
      },
    },
    defaultVariants: {
      variant: 'primary',
      size: 'md',
    },
  }
)

const textVariants = cva('font-medium', {
  variants: {
    variant: {
      primary: 'text-white',
      secondary: 'text-gray-900',
      destructive: 'text-white',
      outline: 'text-gray-900',
    },
    size: {
      sm: 'text-sm',
      md: 'text-base',
      lg: 'text-lg',
    },
  },
})

interface ButtonProps extends PressableProps, VariantProps<typeof buttonVariants> {
  children: React.ReactNode
  onPress?: () => void
}

export function Button({ children, variant, size, className, ...props }: ButtonProps) {
  return (
    <Pressable
      className={buttonVariants({ variant, size, className })}
      {...props}
    >
      <Text className={textVariants({ variant, size })}>{children}</Text>
    </Pressable>
  )
}
```

### 3.4 State Management

```typescript
// packages/stores/src/shopping-list.store.ts
import { create } from 'zustand'
import { persist, createJSONStorage } from 'zustand/middleware'

interface ShoppingListState {
  // State
  currentListId: string | null
  lists: Array<{ id: string; name: string }>

  // Actions
  setCurrentList: (id: string) => void
  addList: (list: { id: string; name: string }) => void
  removeList: (id: string) => void
  clearLists: () => void
}

export const useShoppingListStore = create<ShoppingListState>()(
  persist(
    (set) => ({
      currentListId: null,
      lists: [],

      setCurrentList: (id) => set({ currentListId: id }),

      addList: (list) => set((state) => ({
        lists: [...state.lists, list],
      })),

      removeList: (id) => set((state) => ({
        lists: state.lists.filter((list) => list.id !== id),
        currentListId: state.currentListId === id ? null : state.currentListId,
      })),

      clearLists: () => set({ lists: [], currentListId: null }),
    }),
    {
      name: 'shopping-list-storage',
      storage: createJSONStorage(() => {
        // Platform-specific storage
        if (typeof window !== 'undefined') {
          return localStorage
        }
        // React Native AsyncStorage
        return {
          getItem: async (name) => {
            const AsyncStorage = (await import('@react-native-async-storage/async-storage')).default
            return AsyncStorage.getItem(name)
          },
          setItem: async (name, value) => {
            const AsyncStorage = (await import('@react-native-async-storage/async-storage')).default
            return AsyncStorage.setItem(name, value)
          },
          removeItem: async (name) => {
            const AsyncStorage = (await import('@react-native-async-storage/async-storage')).default
            return AsyncStorage.removeItem(name)
          },
        }
      }),
    }
  )
)
```

---

## 4. Design System & Theming

### 4.1 Design Tokens

```typescript
// packages/tokens/src/colors.ts
export const colors = {
  // Brand colors
  primary: {
    50: '#eff6ff',
    100: '#dbeafe',
    200: '#bfdbfe',
    300: '#93c5fd',
    400: '#60a5fa',
    500: '#3b82f6',
    600: '#2563eb',
    700: '#1d4ed8',
    800: '#1e40af',
    900: '#1e3a8a',
  },

  // Semantic colors
  success: {
    50: '#f0fdf4',
    500: '#22c55e',
    700: '#15803d',
  },

  warning: {
    50: '#fefce8',
    500: '#eab308',
    700: '#a16207',
  },

  error: {
    50: '#fef2f2',
    500: '#ef4444',
    700: '#b91c1c',
  },

  // Neutral colors
  gray: {
    50: '#f9fafb',
    100: '#f3f4f6',
    200: '#e5e7eb',
    300: '#d1d5db',
    400: '#9ca3af',
    500: '#6b7280',
    600: '#4b5563',
    700: '#374151',
    800: '#1f2937',
    900: '#111827',
  },
} as const

// packages/tokens/src/spacing.ts
export const spacing = {
  px: '1px',
  0: '0px',
  0.5: '0.125rem',
  1: '0.25rem',
  1.5: '0.375rem',
  2: '0.5rem',
  2.5: '0.625rem',
  3: '0.75rem',
  3.5: '0.875rem',
  4: '1rem',
  5: '1.25rem',
  6: '1.5rem',
  7: '1.75rem',
  8: '2rem',
  9: '2.25rem',
  10: '2.5rem',
  12: '3rem',
  14: '3.5rem',
  16: '4rem',
  20: '5rem',
  24: '6rem',
  28: '7rem',
  32: '8rem',
  36: '9rem',
  40: '10rem',
} as const

// packages/tokens/src/typography.ts
export const typography = {
  fonts: {
    sans: 'Inter, system-ui, -apple-system, sans-serif',
    mono: 'JetBrains Mono, monospace',
  },

  sizes: {
    xs: '0.75rem',
    sm: '0.875rem',
    base: '1rem',
    lg: '1.125rem',
    xl: '1.25rem',
    '2xl': '1.5rem',
    '3xl': '1.875rem',
    '4xl': '2.25rem',
    '5xl': '3rem',
  },

  weights: {
    normal: '400',
    medium: '500',
    semibold: '600',
    bold: '700',
  },

  lineHeights: {
    tight: '1.25',
    normal: '1.5',
    relaxed: '1.75',
  },
} as const
```

### 4.2 Web Theming (Tailwind + shadcn/ui)

```typescript
// apps/web/tailwind.config.ts
import type { Config } from 'tailwindcss'
import { colors, spacing, typography } from '@kainuguru/tokens'

const config: Config = {
  darkMode: ['class'],
  content: [
    './app/**/*.{js,ts,jsx,tsx,mdx}',
    './components/**/*.{js,ts,jsx,tsx,mdx}',
  ],
  theme: {
    extend: {
      colors: {
        border: 'hsl(var(--border))',
        input: 'hsl(var(--input))',
        ring: 'hsl(var(--ring))',
        background: 'hsl(var(--background))',
        foreground: 'hsl(var(--foreground))',
        primary: {
          DEFAULT: 'hsl(var(--primary))',
          foreground: 'hsl(var(--primary-foreground))',
        },
        secondary: {
          DEFAULT: 'hsl(var(--secondary))',
          foreground: 'hsl(var(--secondary-foreground))',
        },
        destructive: {
          DEFAULT: 'hsl(var(--destructive))',
          foreground: 'hsl(var(--destructive-foreground))',
        },
        muted: {
          DEFAULT: 'hsl(var(--muted))',
          foreground: 'hsl(var(--muted-foreground))',
        },
        accent: {
          DEFAULT: 'hsl(var(--accent))',
          foreground: 'hsl(var(--accent-foreground))',
        },
        // Import brand colors from tokens
        brand: colors.primary,
      },
      spacing: spacing,
      fontFamily: {
        sans: [typography.fonts.sans],
        mono: [typography.fonts.mono],
      },
      fontSize: typography.sizes,
      fontWeight: typography.weights,
      borderRadius: {
        lg: 'var(--radius)',
        md: 'calc(var(--radius) - 2px)',
        sm: 'calc(var(--radius) - 4px)',
      },
    },
  },
  plugins: [require('tailwindcss-animate')],
}

export default config
```

### 4.3 Mobile Theming (Tamagui)

```typescript
// apps/mobile/tamagui.config.ts
import { createTamagui, createTokens } from '@tamagui/core'
import { colors, spacing, typography } from '@kainuguru/tokens'

const tokens = createTokens({
  color: {
    // Map design tokens to Tamagui
    primary: colors.primary[600],
    primaryForeground: '#ffffff',
    secondary: colors.gray[200],
    secondaryForeground: colors.gray[900],
    destructive: colors.error[500],
    destructiveForeground: '#ffffff',

    // Light theme
    background: '#ffffff',
    foreground: colors.gray[900],
    border: colors.gray[200],

    // Dark theme
    backgroundDark: colors.gray[900],
    foregroundDark: '#ffffff',
    borderDark: colors.gray[700],
  },

  space: spacing,

  size: spacing,

  radius: {
    0: 0,
    1: 4,
    2: 8,
    3: 12,
    4: 16,
    full: 9999,
  },

  font: {
    body: { family: typography.fonts.sans },
    mono: { family: typography.fonts.mono },
  },

  fontSize: typography.sizes,

  fontWeight: typography.weights,
})

export const config = createTamagui({
  tokens,
  themes: {
    light: {
      background: tokens.color.background,
      foreground: tokens.color.foreground,
      primary: tokens.color.primary,
      secondary: tokens.color.secondary,
      border: tokens.color.border,
    },
    dark: {
      background: tokens.color.backgroundDark,
      foreground: tokens.color.foregroundDark,
      primary: tokens.color.primary,
      secondary: tokens.color.secondary,
      border: tokens.color.borderDark,
    },
  },
})
```

### 4.4 Mobile Theming (NativeWind)

```javascript
// apps/mobile/tailwind.config.js
const { colors, spacing } = require('@kainuguru/tokens')

module.exports = {
  content: [
    './app/**/*.{js,ts,jsx,tsx}',
    './components/**/*.{js,ts,jsx,tsx}',
  ],
  presets: [require('nativewind/preset')],
  theme: {
    extend: {
      colors: {
        // Direct mapping from tokens
        primary: colors.primary,
        success: colors.success,
        warning: colors.warning,
        error: colors.error,
        gray: colors.gray,
      },
      spacing: spacing,
    },
  },
  plugins: [],
}
```

---

## 5. Development Workflow

### 5.1 Initial Setup

```bash
# 1. Clone and install
git clone <repo>
cd kainuguru-frontend
pnpm install

# 2. Setup environment variables
cp .env.example .env.local
cp apps/web/.env.example apps/web/.env.local
cp apps/mobile/.env.example apps/mobile/.env.local

# 3. Generate GraphQL types
pnpm codegen

# 4. Install shadcn/ui components (web)
cd apps/web
npx shadcn@latest init
npx shadcn@latest add button card dialog form input

# 5. Setup Expo (mobile)
cd apps/mobile
npx expo install
eas init
eas build:configure
```

### 5.2 Development Commands

```json
// Root package.json scripts
{
  "scripts": {
    // Development
    "dev": "turbo dev",
    "dev:web": "turbo dev --filter=web",
    "dev:mobile": "turbo dev --filter=mobile",

    // Building
    "build": "turbo build",
    "build:web": "turbo build --filter=web",
    "build:mobile": "turbo build --filter=mobile",

    // Testing
    "test": "turbo test",
    "test:watch": "turbo test:watch",
    "test:e2e": "turbo test:e2e",

    // Code quality
    "lint": "turbo lint",
    "type-check": "turbo type-check",
    "format": "prettier --write \"**/*.{ts,tsx,md,json}\"",

    // GraphQL
    "codegen": "turbo codegen",
    "codegen:watch": "turbo codegen:watch",

    // Mobile specific
    "ios": "turbo ios --filter=mobile",
    "android": "turbo android --filter=mobile",
    "eas:build": "cd apps/mobile && eas build",
    "eas:submit": "cd apps/mobile && eas submit"
  }
}
```

### 5.3 Git Workflow

```bash
# Branch naming
feature/TICKET-description
bugfix/TICKET-description
hotfix/TICKET-description

# Commit message format
type(scope): description

# Types
feat: New feature
fix: Bug fix
docs: Documentation
style: Formatting
refactor: Code restructuring
test: Tests
chore: Maintenance

# Examples
feat(shopping-list): add barcode scanning
fix(auth): resolve token refresh race condition
docs(readme): update setup instructions
```

---

## 6. Testing Strategy

### 6.1 Unit Tests

```typescript
// packages/utils/src/validation/shopping-list.test.ts
import { describe, it, expect } from 'vitest'
import { shoppingListSchema } from './shopping-list'

describe('Shopping List Validation', () => {
  it('validates correct input', () => {
    const input = {
      name: 'Groceries',
      description: 'Weekly shopping',
    }

    const result = shoppingListSchema.safeParse(input)
    expect(result.success).toBe(true)
  })

  it('rejects empty name', () => {
    const input = {
      name: '',
      description: 'Weekly shopping',
    }

    const result = shoppingListSchema.safeParse(input)
    expect(result.success).toBe(false)
  })
})
```

### 6.2 Integration Tests

```typescript
// apps/web/__tests__/shopping-list.test.tsx
import { render, screen, waitFor } from '@testing-library/react'
import { ShoppingListPage } from '@/app/(dashboard)/lists/page'
import { mockServer } from '@/test/mock-server'

describe('Shopping List Page', () => {
  it('displays shopping lists', async () => {
    mockServer.use(
      graphql.query('GetShoppingLists', (req, res, ctx) => {
        return res(
          ctx.data({
            shoppingLists: {
              edges: [
                { node: { id: '1', name: 'Groceries' } },
                { node: { id: '2', name: 'Hardware' } },
              ],
            },
          })
        )
      })
    )

    render(<ShoppingListPage />)

    await waitFor(() => {
      expect(screen.getByText('Groceries')).toBeInTheDocument()
      expect(screen.getByText('Hardware')).toBeInTheDocument()
    })
  })
})
```

### 6.3 E2E Tests

```typescript
// apps/web/e2e/shopping-list.spec.ts
import { test, expect } from '@playwright/test'

test.describe('Shopping List', () => {
  test('create new shopping list', async ({ page }) => {
    await page.goto('/lists')
    await page.click('text=New List')

    await page.fill('input[name="name"]', 'Test List')
    await page.fill('textarea[name="description"]', 'Test Description')
    await page.click('button[type="submit"]')

    await expect(page).toHaveURL(/\/lists\/[\w-]+/)
    await expect(page.locator('h1')).toContainText('Test List')
  })
})
```

---

## 7. Performance Optimization

### 7.1 Web Optimizations

```typescript
// Dynamic imports for code splitting
const BarcodeScannerModal = dynamic(
  () => import('@/components/features/barcode-scanner'),
  {
    loading: () => <Skeleton className="h-96 w-full" />,
    ssr: false
  }
)

// Image optimization
import Image from 'next/image'

<Image
  src="/product.jpg"
  alt="Product"
  width={300}
  height={200}
  placeholder="blur"
  blurDataURL={blurDataUrl}
  priority={isAboveFold}
/>

// Prefetching
import { useRouter } from 'next/navigation'

const router = useRouter()
router.prefetch('/lists/123')
```

### 7.2 Mobile Optimizations

```typescript
// React Native List optimization
import { FlashList } from '@shopify/flash-list'

<FlashList
  data={items}
  renderItem={renderItem}
  estimatedItemSize={80}
  keyExtractor={(item) => item.id}
/>

// Image caching
import FastImage from 'react-native-fast-image'

<FastImage
  source={{
    uri: imageUrl,
    priority: FastImage.priority.normal,
    cache: FastImage.cacheControl.immutable,
  }}
  style={{ width: 200, height: 200 }}
  resizeMode={FastImage.resizeMode.cover}
/>
```

---

## 8. Deployment

### 8.1 Web Deployment (Vercel)

```yaml
# .github/workflows/deploy-web.yml
name: Deploy Web

on:
  push:
    branches: [main]
    paths:
      - 'apps/web/**'
      - 'packages/**'

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v3
      - uses: actions/setup-node@v4

      - run: pnpm install --frozen-lockfile
      - run: pnpm codegen
      - run: pnpm build --filter=web

      - uses: amondnet/vercel-action@v25
        with:
          vercel-token: ${{ secrets.VERCEL_TOKEN }}
          vercel-org-id: ${{ secrets.VERCEL_ORG_ID }}
          vercel-project-id: ${{ secrets.VERCEL_PROJECT_ID }}
          vercel-args: '--prod'
```

### 8.2 Mobile Deployment (EAS)

```bash
# Build for production
eas build --platform all --profile production

# Submit to stores
eas submit --platform ios --latest
eas submit --platform android --latest

# OTA update
eas update --branch production --message "Bug fixes"
```

---

## 9. Troubleshooting Guide

### Common Issues

| Issue | Solution |
|-------|----------|
| **Metro bundler fails** | Clear cache: `npx expo start --clear` |
| **Type errors after codegen** | Run `pnpm type-check` from root |
| **shadcn/ui component not found** | Install: `npx shadcn@latest add [component]` |
| **Tamagui styles not applying** | Check `tamagui.config.ts` imports |
| **NativeWind classes not working** | Verify `tailwind.config.js` content paths |
| **GraphQL types outdated** | Run `pnpm codegen` |
| **Turbo cache issues** | Clear: `pnpm turbo clean` |

---

## 10. Migration Checklist

When another agent takes over, they should:

1. [ ] Read this entire document
2. [ ] Clone the repository
3. [ ] Install dependencies with `pnpm install`
4. [ ] Set up environment variables
5. [ ] Run `pnpm codegen` to generate GraphQL types
6. [ ] Verify Go backend is running at configured URL
7. [ ] Run `pnpm dev` to start all applications
8. [ ] Test web at http://localhost:3000
9. [ ] Test mobile with Expo Go or simulator
10. [ ] Review the folder structure
11. [ ] Understand the API integration pattern (Go handles all business logic)
12. [ ] Review authentication flows for both platforms
13. [ ] Familiarize with the UI component patterns
14. [ ] Check CI/CD pipelines in GitHub Actions

---

## Key Architectural Decisions

1. **Go Backend Owns Business Logic**: Frontend is purely presentational
2. **Platform-Specific UI**: Best experience on each platform
3. **Shared Business Logic**: 70-95% code reuse through packages
4. **Type Safety**: End-to-end TypeScript with GraphQL codegen
5. **Modern Stack**: React 19, Next.js 15, Expo SDK 54+

---

## Support & Resources

- Next.js Documentation: https://nextjs.org/docs
- Expo Documentation: https://docs.expo.dev
- shadcn/ui Components: https://ui.shadcn.com
- Tamagui Documentation: https://tamagui.dev
- NativeWind Documentation: https://nativewind.dev
- Turborepo Documentation: https://turbo.build/repo/docs

---

This document should be kept up to date as the architecture evolves.