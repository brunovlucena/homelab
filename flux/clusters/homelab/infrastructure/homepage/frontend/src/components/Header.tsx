import React, { useState } from 'react'
import { Link } from 'react-router-dom'
import { Search, FileText, Linkedin, Github, Mail, Menu, X } from 'lucide-react'
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
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

  const toggleMobileMenu = () => {
    setIsMobileMenuOpen(!isMobileMenuOpen);
  };

  const closeMobileMenu = () => {
    setIsMobileMenuOpen(false);
  };
  
  const navigationLinks: HeaderLink[] = [
    {
      to: '/?section=projects',
      icon: <Search className="header-link-icon" />,
      label: 'Search',
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
    const handleClick = (e?: React.MouseEvent) => {
      closeMobileMenu();
      if (link.onClick) {
        e?.preventDefault();
        link.onClick();
      }
    };

    const linkProps = {
      className: 'header-link',
      ...(link.external && {
        target: '_blank',
        rel: 'noopener noreferrer',
      }),
      onClick: handleClick,
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
            <span className="logo-text">🚀 Bruno Lucena</span>
          </Link>
          
          {/* Homelab Link */}
          <Link to="/?section=projects" className="homelab-link">
            Homelab
          </Link>
        </div>

        {/* Mobile Menu Toggle */}
        <button
          className={`mobile-menu-toggle ${isMobileMenuOpen ? 'active' : ''}`}
          onClick={toggleMobileMenu}
          aria-label="Toggle menu"
          aria-expanded={isMobileMenuOpen}
        >
          {isMobileMenuOpen ? <X size={24} /> : <Menu size={24} />}
        </button>
        
        {/* Navigation */}
        <nav className={`header-nav ${isMobileMenuOpen ? 'nav-open' : ''}`}>
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
