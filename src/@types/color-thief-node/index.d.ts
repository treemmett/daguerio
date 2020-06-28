declare module 'color-thief-node' {
  export function getColorFromURL(url: string): Promise<[number]>;
  export function getPaletteFromURL(url: string): Promise<[[number]]>;
}
