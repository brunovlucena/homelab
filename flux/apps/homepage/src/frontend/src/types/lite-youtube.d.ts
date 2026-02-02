// Type declarations for lite-youtube-embed web component
declare namespace JSX {
  interface IntrinsicElements {
    'lite-youtube': React.DetailedHTMLProps<
      React.HTMLAttributes<HTMLElement> & {
        videoid?: string | null
        playlabel?: string
        params?: string
        style?: React.CSSProperties
      },
      HTMLElement
    >
  }
}
