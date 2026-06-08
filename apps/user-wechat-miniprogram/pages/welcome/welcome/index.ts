import { switchTabRoute } from "../../../utils/navigation";

const FULL_TITLE = "欢迎来到悦享e食";
const TITLE_LINE = "悦享e食";

let typewriterTimer: number | null = null;
let transitionTimers: number[] = [];

function clearWelcomeTimers() {
  if (typewriterTimer !== null) {
    clearInterval(typewriterTimer);
    typewriterTimer = null;
  }

  transitionTimers.forEach((timer) => clearTimeout(timer));
  transitionTimers = [];
}

Page({
  data: {
    showInitialContent: true,
    showFinalContent: false,
    showTitle: false,
    showSubtitle: false,
    showCard: false,
    displayedTitle: "",
    displayedTitleLine2: "",
    typewriterIndex: 0
  },

  onLoad() {
    clearWelcomeTimers();

    if (this.isUserLoggedIn()) {
      this.goHome();
      return;
    }

    this.startAnimation();
  },

  onUnload() {
    clearWelcomeTimers();
  },

  isUserLoggedIn() {
    const authMode = String(wx.getStorageSync("authMode") || "");
    const authToken = String(wx.getStorageSync("authToken") || "");
    const legacyToken = String(wx.getStorageSync("token") || "");
    const legacyRefreshToken = String(wx.getStorageSync("refreshToken") || "");

    if (authMode === "guest") {
      return false;
    }

    return Boolean(authToken || (authMode === "user" && legacyToken && legacyRefreshToken));
  },

  startAnimation() {
    transitionTimers.push(
      setTimeout(() => {
        this.startTypewriter();
      }, 1000)
    );

    transitionTimers.push(
      setTimeout(() => {
        this.setData({
          displayedTitle: FULL_TITLE,
          showInitialContent: false,
          showFinalContent: true,
          showSubtitle: true
        });
      }, 2500)
    );

    transitionTimers.push(
      setTimeout(() => {
        this.setData({
          showCard: true
        });

        transitionTimers.push(
          setTimeout(() => {
            if (String(wx.getStorageSync("authMode") || "") === "user") {
              this.goHome();
            }
          }, 1000)
        );
      }, 3700)
    );
  },

  startTypewriter() {
    this.setData({
      showTitle: true,
      typewriterIndex: 0,
      displayedTitle: "",
      displayedTitleLine2: ""
    });

    typewriterTimer = setInterval(() => {
      const typewriterIndex = Number(this.data.typewriterIndex || 0);

      if (typewriterIndex < TITLE_LINE.length) {
        this.setData({
          displayedTitleLine2: TITLE_LINE.slice(0, typewriterIndex + 1),
          typewriterIndex: typewriterIndex + 1
        });
        return;
      }

      this.setData({
        displayedTitle: FULL_TITLE
      });
      if (typewriterTimer !== null) {
        clearInterval(typewriterTimer);
        typewriterTimer = null;
      }
    }, 150);
  },

  goLogin() {
    wx.navigateTo({
      url: "/pages/auth/login/index",
      fail() {
        wx.showToast({
          title: "登录页下一步接入",
          icon: "none"
        });
      }
    });
  },

  goGuest() {
    wx.setStorageSync("authMode", "guest");
    this.goHome();
  },

  goHome() {
    switchTabRoute("/pages/index/index");
  }
});
