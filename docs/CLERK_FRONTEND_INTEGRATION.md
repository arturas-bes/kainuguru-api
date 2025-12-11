# Clerk Frontend Integration Plan

## Overview
This document outlines how to integrate Clerk authentication in the frontend (Next.js) to work with the kainuguru-api backend.

## Prerequisites
- Clerk account at https://clerk.com/
- Clerk application created with Lithuanian locale support
- Environment variables from Clerk dashboard

## 1. Install Clerk SDK

```bash
npm install @clerk/nextjs
```

## 2. Environment Variables

Add to your `.env.local`:

```env
NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=pk_test_...
CLERK_SECRET_KEY=sk_test_...
NEXT_PUBLIC_CLERK_SIGN_IN_URL=/sign-in
NEXT_PUBLIC_CLERK_SIGN_UP_URL=/sign-up
NEXT_PUBLIC_CLERK_AFTER_SIGN_IN_URL=/
NEXT_PUBLIC_CLERK_AFTER_SIGN_UP_URL=/
```

## 3. Configure ClerkProvider

Update your `app/layout.tsx`:

```tsx
import { ClerkProvider } from '@clerk/nextjs';
import { ltLT } from '@clerk/localizations';

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <ClerkProvider localization={ltLT}>
      <html lang="lt">
        <body>{children}</body>
      </html>
    </ClerkProvider>
  );
}
```

## 4. Middleware for Protected Routes

Create `middleware.ts` in your project root:

```ts
import { clerkMiddleware, createRouteMatcher } from '@clerk/nextjs/server';

// Define public routes that don't require authentication
const isPublicRoute = createRouteMatcher([
  '/',
  '/sign-in(.*)',
  '/sign-up(.*)',
  '/api/webhook(.*)',
  '/products(.*)',
  '/flyers(.*)',
  '/stores(.*)',
]);

export default clerkMiddleware(async (auth, request) => {
  if (!isPublicRoute(request)) {
    await auth.protect();
  }
});

export const config = {
  matcher: [
    '/((?!_next|[^?]*\\.(?:html?|css|js(?!on)|jpe?g|webp|png|gif|svg|ttf|woff2?|ico|csv|docx?|xlsx?|zip|webmanifest)).*)',
    '/(api|trpc)(.*)',
  ],
};
```

## 5. Authentication Pages

### Sign In Page (`app/sign-in/[[...sign-in]]/page.tsx`):

```tsx
import { SignIn } from '@clerk/nextjs';

export default function SignInPage() {
  return (
    <div className="flex min-h-screen items-center justify-center">
      <SignIn
        appearance={{
          elements: {
            rootBox: 'mx-auto',
            card: 'shadow-lg',
          }
        }}
      />
    </div>
  );
}
```

### Sign Up Page (`app/sign-up/[[...sign-up]]/page.tsx`):

```tsx
import { SignUp } from '@clerk/nextjs';

export default function SignUpPage() {
  return (
    <div className="flex min-h-screen items-center justify-center">
      <SignUp
        appearance={{
          elements: {
            rootBox: 'mx-auto',
            card: 'shadow-lg',
          }
        }}
      />
    </div>
  );
}
```

## 6. User Button Component

Add to your header/navbar:

```tsx
import { UserButton, SignedIn, SignedOut, SignInButton } from '@clerk/nextjs';

export function Header() {
  return (
    <header>
      <nav>
        {/* Your nav items */}
        <SignedIn>
          <UserButton afterSignOutUrl="/" />
        </SignedIn>
        <SignedOut>
          <SignInButton mode="modal">
            <button>Prisijungti</button>
          </SignInButton>
        </SignedOut>
      </nav>
    </header>
  );
}
```

## 7. Making Authenticated API Requests

### Using `useAuth` hook with fetch:

```tsx
import { useAuth } from '@clerk/nextjs';

export function useAuthenticatedFetch() {
  const { getToken } = useAuth();

  const fetchWithAuth = async (url: string, options: RequestInit = {}) => {
    const token = await getToken();

    return fetch(url, {
      ...options,
      headers: {
        ...options.headers,
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
    });
  };

  return { fetchWithAuth };
}
```

### With GraphQL (Apollo Client):

```tsx
import { ApolloClient, InMemoryCache, createHttpLink } from '@apollo/client';
import { setContext } from '@apollo/client/link/context';
import { useAuth } from '@clerk/nextjs';

export function useApolloClient() {
  const { getToken } = useAuth();

  const httpLink = createHttpLink({
    uri: process.env.NEXT_PUBLIC_GRAPHQL_URL || 'http://localhost:8080/graphql',
  });

  const authLink = setContext(async (_, { headers }) => {
    const token = await getToken();
    return {
      headers: {
        ...headers,
        authorization: token ? `Bearer ${token}` : '',
      },
    };
  });

  return new ApolloClient({
    link: authLink.concat(httpLink),
    cache: new InMemoryCache(),
  });
}
```

### With TanStack Query:

```tsx
import { useAuth } from '@clerk/nextjs';
import { useQuery } from '@tanstack/react-query';

export function useMe() {
  const { getToken, isSignedIn } = useAuth();

  return useQuery({
    queryKey: ['me'],
    queryFn: async () => {
      const token = await getToken();
      const response = await fetch('/api/graphql', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${token}`,
        },
        body: JSON.stringify({
          query: `
            query Me {
              me {
                id
                email
                fullName
                emailVerified
                preferredLanguage
              }
            }
          `,
        }),
      });
      const data = await response.json();
      return data.data.me;
    },
    enabled: isSignedIn,
  });
}
```

## 8. Protected Components

```tsx
import { useAuth } from '@clerk/nextjs';
import { redirect } from 'next/navigation';

export function ProtectedComponent({ children }: { children: React.ReactNode }) {
  const { isLoaded, isSignedIn } = useAuth();

  if (!isLoaded) {
    return <div>Loading...</div>;
  }

  if (!isSignedIn) {
    redirect('/sign-in');
  }

  return <>{children}</>;
}
```

## 9. Server-Side Authentication (App Router)

```tsx
import { auth } from '@clerk/nextjs/server';

export default async function ProtectedPage() {
  const { userId, getToken } = await auth();

  if (!userId) {
    redirect('/sign-in');
  }

  // Make authenticated API call
  const token = await getToken();
  const response = await fetch('http://localhost:8080/graphql', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify({
      query: `query { me { id email } }`,
    }),
  });

  const data = await response.json();

  return <div>Welcome, {data.data.me.email}</div>;
}
```

## 10. Clerk Dashboard Configuration

### Recommended Settings:

1. **Authentication Methods**:
   - Email + Password
   - Google OAuth
   - Facebook OAuth (optional)
   - Apple OAuth (optional)
   - Phone number (optional)

2. **Multi-factor Authentication**:
   - Enable TOTP (authenticator apps)
   - SMS verification (optional)

3. **Session Settings**:
   - Session token lifetime: 1 hour
   - Multi-session: Allow (for multiple devices)

4. **Localization**:
   - Default language: Lithuanian (lt-LT)
   - Enable: English as fallback

## 11. Backend Configuration

To enable Clerk on the backend, set these environment variables:

```env
CLERK_ENABLED=true
CLERK_SECRET_KEY=sk_test_... # Get from Clerk Dashboard > API Keys
CLERK_PUBLISHABLE_KEY=pk_test_... # Get from Clerk Dashboard > API Keys
```

## 12. Migration Notes

### From Previous Auth System:
- Remove `/api/auth/*` routes (register, login, logout)
- Remove `refreshToken` mutation usage
- Update all authenticated requests to use Clerk's `getToken()`
- Update auth state management to use Clerk's `useAuth()`
- Remove local session/cookie handling

### What Still Works:
- `me` GraphQL query - returns the authenticated user
- Shopping list operations - automatically linked to Clerk user
- All existing GraphQL queries/mutations that require auth

## 13. Testing Checklist

- [ ] Sign up with email
- [ ] Sign in with email
- [ ] Sign in with Google
- [ ] Sign out
- [ ] Protected pages redirect to sign-in
- [ ] API requests include Bearer token
- [ ] `me` query returns user data
- [ ] Shopping list operations work
- [ ] Multi-factor authentication (if enabled)
- [ ] Password reset flow
