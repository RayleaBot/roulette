import vue from "@vitejs/plugin-vue";
import { defineConfig } from "vitest/config";

export default defineConfig({
  base: "./",
  plugins: [vue()],
  resolve: { dedupe: ["vue"] },
  build: { emptyOutDir: true },
  test: {
    environment: "jsdom",
    environmentOptions: {
      jsdom: {
        url: "http://localhost/plugin-ui/raylea.roulette/index.html?page=settings",
      },
    },
    include: ["tests/**/*.test.ts"],
  },
});
