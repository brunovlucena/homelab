import React from 'react'
import { Link } from 'react-router-dom'
import { FileText, BookOpen, Linkedin, Github, Mail } from 'lucide-react'
import { useChatbot } from '../contexts/ChatbotContext'

// =============================================================================
// 📋 TYPES
// =============================================================================

interface HeaderLink {
  to: string
  icon: React.ReactNode
  label: string
  external?: boolean
  onClick?: () => void
}

// =============================================================================
// 🎯 HEADER COMPONENT
// =============================================================================

const Header: React.FC = () => {
  const { openChatbot } = useChatbot();
  
  const navigationLinks: HeaderLink[] = [
    {
      to: '/blog',
      icon: <BookOpen className="header-link-icon" />,
      label: 'Blog',
    },
    {
      to: '/resume',
      icon: <FileText className="header-link-icon" />,
      label: 'Resume',
    },
    {
      to: 'https://www.linkedin.com/in/bvlucena',
      icon: <Linkedin className="header-link-icon" />,
      label: 'LinkedIn',
      external: true,
    },
    {
      to: 'https://github.com/brunovlucena',
      icon: <Github className="header-link-icon" />,
      label: 'GitHub',
      external: true,
    },
    {
      to: '#',
      icon: <Mail className="header-link-icon" />,
      label: 'Contact',
      onClick: openChatbot,
    },
  ]

  const renderLink = (link: HeaderLink) => {
    const linkProps = {
      className: 'header-link',
      ...(link.external && {
        target: '_blank',
        rel: 'noopener noreferrer',
      }),
      ...(link.onClick && {
        onClick: link.onClick,
      }),
    }

    if (link.external) {
      return (
        <a key={link.label} href={link.to} {...linkProps}>
          {link.icon}
          <span>{link.label}</span>
        </a>
      )
    }

    if (link.onClick) {
      return (
        <button key={link.label} {...linkProps} type="button">
          {link.icon}
          <span>{link.label}</span>
        </button>
      )
    }

    return (
      <Link key={link.label} to={link.to} {...linkProps}>
        {link.icon}
        <span>{link.label}</span>
      </Link>
    )
  }

  return (
    <header className="header">
      <div className="header-container">
        {/* Logo */}
        <div className="header-brand">
          <Link to="/" className="logo">
            <span className="logo-text">Bruno Lucena</span>
          </Link>
        </div>
        
        {/* Navigation */}
        <nav className="header-nav">
          <ul className="nav-menu">
            {navigationLinks.map((link) => (
              <li key={link.label}>
                {renderLink(link)}
              </li>
            ))}
          </ul>
        </nav>
      </div>
    </header>
  )
}

export default Header
