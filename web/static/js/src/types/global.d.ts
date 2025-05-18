// Global TypeScript declarations

// Declare types for global objects
declare const __APP_VERSION__: string;
declare const __DEV__: boolean;

// Declare types for modules without type definitions
declare module '*.css' {
  const content: { [className: string]: string };
  export default content;
}

declare module '*.svg' {
  import * as React from 'react';
  export const ReactComponent: React.FunctionComponent<React.SVGProps<SVGSVGElement>>;
  const src: string;
  export default src;
}

// Extend the Window interface to include any global browser APIs or custom properties
interface Window {
  // Add any global browser APIs or custom properties here
  htmx?: {
    defineExtension?: (name: string, extension: any) => void;
    process?: (element: HTMLElement) => void;
    on?: (event: string, handler: (event: CustomEvent) => void) => void;
    off?: (event: string, handler: (event: CustomEvent) => void) => void;
    trigger?: (element: HTMLElement, event: string, detail?: any) => void;
    find?: (selector: string) => HTMLElement | null;
    findAll?: (selector: string) => NodeListOf<HTMLElement>;
    ajax?: (method: string, url: string, config: any) => void;
  };
  
  // WebSocket client for Go Sentinel
  goSentinelWebSocket?: {
    connect: (url: string) => void;
    disconnect: () => void;
    send: (data: any) => void;
    on: (event: string, callback: (data: any) => void) => void;
    off: (event: string, callback: (data: any) => void) => void;
  };

  // Coverage visualization functions
  showFileDetails?: (filePath?: string) => void;
  getCoverageClass?: (percentage: number) => string;
  goToPage?: (page: number, filter?: string, search?: string) => void;
}

// Declare types for any global variables or functions
declare const toast: {
  success: (message: string, options?: any) => void;
  error: (message: string, options?: any) => void;
  info: (message: string, options?: any) => void;
  warning: (message: string, options?: any) => void;
};

// Declare types for any custom events
declare namespace CustomEventMap {
  interface WebSocketMessageEvent extends CustomEvent {
    detail: {
      type: string;
      payload: any;
    };
  }
  // Add more custom event types as needed
}

declare global {
  // Extend the global EventTarget interface to include our custom events
  interface WindowEventMap extends CustomEventMap {}
  interface HTMLElementEventMap extends CustomEventMap {}
  
  // Add any global utility types here
  type Nullable<T> = T | null;
  type Optional<T> = T | undefined;
  type Dictionary<T> = Record<string, T>;
  
  // Add any global utility functions here
  function debounce<T extends (...args: any[]) => any>(
    func: T,
    wait: number,
    immediate?: boolean
  ): (...args: Parameters<T>) => void;
}
