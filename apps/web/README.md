# Arc Privacy Platform - Frontend

Modern, enterprise-grade frontend for Arc Consent Management and Privacy Compliance Platform.

## 🚀 Quick Start

```bash
# Install dependencies
npm install

# Run development server
npm run dev

# Build for production
npm run build

# Start production server
npm start
```

Visit `http://localhost:3000` to see the application.

## 🎨 Design System

### Color Palette
- **Primary Purple**: `#6D28D9` - Brand color for buttons, links, emphasis
- **Purple Hover**: `#5B21B6` - Hover states
- **Purple Light**: `#EDE9FE` - Backgrounds, highlights
- **Success**: `#059669` - Positive actions
- **Warning**: `#D97706` - Caution states  
- **Error**: `#DC2626` - Errors, destructive actions
- **Info**: `#0284C7` - Informational elements

### Typography
- **Font Family**: Inter (sans-serif), Fira Code (monospace)
- **Sizes**: H1 (32px), H2 (24px), H3 (20px), H4 (18px), Body (16px), Small (14px), Tiny (12px)

### Spacing
- XS: 4px, SM: 8px, MD: 16px, LG: 24px, XL: 32px, 2XL: 48px, 3XL: 64px

## 📁 Project Structure

```
apps/web/
├── app/                    # Next.js App Router
│   ├── login/             # Login page
│   ├── signup/            # Signup pages
│   ├── dashboard/         # Protected dashboard routes
│   ├── globals.css        # Global styles with Tailwind
│   ├── layout.tsx         # Root layout
│   └── page.tsx           # Homepage
├── components/
│   ├── ui/                # Reusable UI components
│   │   ├── Button.tsx
│   │   ├── Input.tsx
│   │   └── Card.tsx
│   ├── layout/            # Layout components (coming soon)
│   └── features/          # Feature-specific components (coming soon)
├── lib/
│   └── utils.ts           # Utility functions
├── public/                # Static assets
└── tailwind.config.ts     # Tailwind configuration
```

## 🧩 Components

### Button
```tsx
import { Button } from '@/components/ui/Button';

<Button variant="primary">Click me</Button>
<Button variant="secondary" size="lg">Large Button</Button>
<Button variant="danger" loading>Processing...</Button>
```

**Variants**: `primary`, `secondary`, `tertiary`, `danger`
**Sizes**: `sm`, `md`, `lg`
**Props**: `loading`, `fullWidth`, `disabled`

### Input
```tsx
import { Input } from '@/components/ui/Input';

<Input 
  label="Email" 
  type="email" 
  error="Invalid email"
  helperText="We'll never share your email"
  required
/>
```

### Card
```tsx
import { Card, CardHeader, CardTitle, CardContent } from '@/components/ui/Card';

<Card>
  <CardHeader>
    <CardTitle>Card Title</CardTitle>
  </CardHeader>
  <CardContent>
    Content goes here
  </CardContent>
</Card>
```

## ♿ Accessibility (WCAG 2.1 AA)

All components follow WCAG 2.1 AA guidelines:

- **Keyboard Navigation**: All interactive elements accessible via keyboard
- **Focus Indicators**: Visible focus rings on all focusable elements
- **ARIA Labels**: Proper ARIA attributes for screen readers
- **Color Contrast**: Minimum 4.5:1 ratio for normal text, 3:1 for large text
- **Form Labels**: All inputs properly labeled with `for` attributes
- **Error Announcements**: Error messages announced via `aria-live`
- **Semantic HTML**: Proper heading hierarchy and landmark elements

## 🔐 Authentication

### Login Page (`/login`)
- Email/password authentication
- User type toggle (Data Principal / Fiduciary)
- SSO integration (Google Workspace, Microsoft 365)
- Form validation with Zod
- Loading states
- Responsive design

## 🎯 Tech Stack

- **Framework**: Next.js 14 (App Router)
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **Forms**: React Hook Form + Zod
- **HTTP Client**: Axios
- **Utilities**: clsx, tailwind-merge

## 📊 Features Implemented

### ✅ Phase 1 - Foundation (Completed)
- [x] Project setup with Next.js 14
- [x] Tailwind CSS configuration with purple theme
- [x] Design system foundation
- [x] Core UI components (Button, Input, Card, Badge, Select)
- [x] Login page with form validation
- [x] SSO buttons (Google Workspace, Microsoft 365)
- [x] Homepage with branding
- [x] Data Principal signup with guardian verification
- [x] Fiduciary signup with organization details
- [x] Password reset flow (forgot + reset pages)
- [x] Password strength indicator
- [x] Multi-step form component
- [x] Step indicator component

### 🚧 Phase 1 - In Progress
- [ ] Protected dashboard layout
- [ ] Profile management
- [ ] Purpose management
- [ ] Consent form builder

### 📋 Phase 2 - Planned
- [ ] DSR management
- [ ] Breach notifications
- [ ] Audit logs viewer
- [ ] Grievance system
- [ ] API key management
- [ ] SDK configuration

## 🌐 Environment Variables

Create a `.env.local` file:

```bash
NEXT_PUBLIC_API_URL=http://localhost:8080
```

## 📝 Code Style

- Use TypeScript for all files
- Use named exports for components
- Use `forwardRef` for components that need ref support
- Use `cn()` utility for conditional className merging
- Follow accessibility best practices
- Add proper TypeScript types

## 🧪 Testing (Coming Soon)

```bash
npm run test          # Run tests
npm run test:watch    # Watch mode
npm run test:coverage # Coverage report
```

## 🚀 Deployment (Coming Soon)

```bash
npm run build
npm start
```

## 📚 Documentation

- [Design System Guide](./docs/design-system.md) (TBD)
- [Component API](./docs/components.md) (TBD)
- [Accessibility Guide](./docs/accessibility.md) (TBD)

## 🤝 Contributing

1. Follow the existing code style
2. Ensure WCAG 2.1 AA compliance
3. Add TypeScript types
4. Test across browsers
5. Update documentation

## 📄 License

Proprietary - Arc Privacy Platform
