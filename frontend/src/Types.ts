export interface InspectResponse {
  url: string;
  status_code: number;
  status_msg: string;
  html_version: string;
  page_title: string;
  headings: Headings;
  login_field_count: number;
  links: Link[];
  accessible_link_count: number;
  inaccessible_link_count: number;
  not_analysed_link_count: number;
  total_link_count: number;
  external_link_count: number;
  internal_link_count: number;
}

export interface Headings {
  h1: string[];
  h2: string[];
  h3: string[];
  h4: string[];
  h5: string[];
  h6: string[];
}

export interface Link {
  url: string;
  text: string;
  type: string;
  status_code: number;
}

