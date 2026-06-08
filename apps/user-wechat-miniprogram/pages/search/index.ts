import { searchCatalog } from "../../utils/api";

type SearchResultCard = {
  id: string;
  type: string;
  title: string;
  subtitle: string;
  imageUrl: string;
  priceText: string;
  distanceText: string;
  badge: string;
  buttonText: string;
  route: string;
  typeClass: string;
};

Page({
  data: {
    keyword: "",
    statusText: "实时搜索",
    loading: false,
    emptyText: "",
    hasKeyword: false,
    selectedCategory: "all",
    categories: [
      { id: "all", title: "全部", active: "active" },
      { id: "shop", title: "商家", active: "" },
      { id: "product", title: "菜品", active: "" },
      { id: "groupbuy", title: "团购", active: "" },
      { id: "medicine", title: "买药", active: "" },
      { id: "errand", title: "跑腿", active: "" }
    ],
    suggestions: [] as string[],
    totalText: "正在读取后端搜索结果",
    results: [] as SearchResultCard[]
  },
  searchTimer: 0 as number,
  onLoad(query) {
    const keyword = String(query?.keyword || "").trim();
    this.setData({ keyword, hasKeyword: Boolean(keyword) });
    if (keyword) {
      this.handleSearch();
      return;
    }
    this.loadGuess();
  },
  async loadGuess() {
    this.setData({
      hasKeyword: false,
      loading: true,
      statusText: "猜你喜欢",
      emptyText: "",
      totalText: "猜你喜欢",
      results: []
    });
    try {
      const catalog = await searchCatalog("", "all") as Record<string, unknown>;
      const suggestions = normalizeSuggestions(catalog.suggestions);
      this.setData({
        suggestions,
        emptyText: suggestions.length ? "" : "后端暂无猜你喜欢。",
        statusText: "猜你喜欢",
        loading: false
      });
    } catch (error) {
      this.setData({
        suggestions: [],
        emptyText: String((error as any)?.message || "搜索推荐接口连接失败，请稍后重试。"),
        statusText: "接口异常",
        loading: false
      });
    }
  },
  async handleSearch() {
    if (this.searchTimer) {
      clearTimeout(this.searchTimer);
      this.searchTimer = 0;
    }
    const keyword = String(this.data.keyword || "").trim();
    if (!keyword) {
      this.loadGuess();
      return;
    }
    this.setData({
      hasKeyword: true,
      loading: true,
      statusText: "搜索中",
      emptyText: "",
      keyword
    });
    try {
      const catalog = await searchCatalog(keyword, this.data.selectedCategory) as Record<string, unknown>;
      const results = (catalog.results as Array<Record<string, unknown>> || []).map(resultFromApi);
      this.setData({
        suggestions: normalizeSuggestions(catalog.suggestions),
        totalText: `找到 ${catalog.total || results.length} 个相关结果`,
        results,
        emptyText: results.length ? "" : emptyTextForSearch(keyword),
        statusText: "实时结果",
        loading: false
      });
    } catch (error) {
      this.setData({
        totalText: "搜索接口暂不可用",
        results: [],
        emptyText: String((error as any)?.message || "搜索接口连接失败，请稍后重试。"),
        statusText: "接口异常",
        loading: false
      });
    }
  },
  handleKeywordInput(event) {
    const keyword = String(event.detail.value || "").trim();
    this.setData({ keyword, hasKeyword: Boolean(keyword) });
    if (this.searchTimer) {
      clearTimeout(this.searchTimer);
      this.searchTimer = 0;
    }
    if (!keyword) {
      this.loadGuess();
      return;
    }
    this.searchTimer = setTimeout(() => {
      this.handleSearch();
    }, 360) as unknown as number;
  },
  handleClearKeyword() {
    if (this.searchTimer) {
      clearTimeout(this.searchTimer);
      this.searchTimer = 0;
    }
    this.setData({ keyword: "" });
    this.loadGuess();
  },
  handleCategoryTap(event) {
    const selectedCategory = String(event.currentTarget.dataset.id || "all");
    this.setData({
      selectedCategory,
      categories: this.data.categories.map((item) => ({ ...item, active: item.id === selectedCategory ? "active" : "" }))
    });
    this.handleSearch();
  },
  handleSuggestionTap(event) {
    this.setData({ keyword: String(event.currentTarget.dataset.keyword || "") });
    this.handleSearch();
  },
  handleResultTap(event) {
    wx.navigateTo({ url: String(event.currentTarget.dataset.route || "/pages/index/index") });
  },
  handleBack() {
    wx.navigateBack({ fail: () => wx.switchTab({ url: "/pages/index/index" }) });
  }
});

function resultFromApi(item: Record<string, unknown>) {
  const priceFen = Number(item.price_fen || 0);
  const type = String(item.type || "shop");
  return {
    id: String(item.id || Date.now()),
    type,
    title: String(item.title || ""),
    subtitle: String(item.subtitle || ""),
    imageUrl: String(item.image_url || ""),
    priceText: priceFen > 0 ? `¥${(priceFen / 100).toFixed(2)}` : "",
    distanceText: String(item.distance_text || ""),
    badge: String(item.badge || typeText(type)),
    buttonText: String(item.button_text || "去使用"),
    route: String(item.route || "/pages/index/index"),
    typeClass: type
  };
}

function normalizeSuggestions(value: unknown) {
  if (!Array.isArray(value)) return [];
  return value.map((item) => String(item || "").trim()).filter(Boolean);
}

function emptyTextForSearch(keyword: string) {
  if (keyword) return "后端没有返回相关结果，换个关键词试试。";
  return "后端暂时没有可展示的搜索结果。";
}

function typeText(type: string) {
  switch (type) {
    case "product":
      return "菜品";
    case "groupbuy":
      return "团购";
    case "medicine":
      return "药店";
    case "errand":
      return "跑腿";
    default:
      return "商家";
  }
}
