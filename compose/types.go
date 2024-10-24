package compose

type GetChannelResonse struct {
	Result struct {
		Channel struct {
			ID                           string `json:"id"`
			URL                          string `json:"url"`
			Name                         string `json:"name"`
			Description                  string `json:"description"`
			DescriptionMentions          []int  `json:"descriptionMentions"`
			DescriptionMentionsPositions []int  `json:"descriptionMentionsPositions"`
			ImageURL                     string `json:"imageUrl"`
			HeaderImageURL               string `json:"headerImageUrl"`
			LeadFid                      int    `json:"leadFid"`
			ModeratorFids                []int  `json:"moderatorFids"`
			CreatedAt                    int64  `json:"createdAt"`
			FollowerCount                int    `json:"followerCount"`
			MemberCount                  int    `json:"memberCount"`
			PinnedCastHash               string `json:"pinnedCastHash"`
			PublicCasting                bool   `json:"publicCasting"`
			ExternalLink                 struct {
				Title string `json:"title"`
				URL   string `json:"url"`
			} `json:"externalLink"`
		} `json:"channel"`
	} `json:"result"`
}

type CastParentID struct {
	Fid  int    `json:"fid"`
	Hash string `json:"hash"`
}

type CastAddBody struct {
	EmbedsDeprecated  []interface{} `json:"embedsDeprecated"`
	Mentions          []interface{} `json:"mentions"`
	ParentCastId      *CastParentID `json:"parentCastId"`
	Text              string        `json:"text"`
	MentionsPositions []interface{} `json:"mentionsPositions"`
	Embeds            []interface{} `json:"embeds"`
}

type MessageData struct {
	Type        string      `json:"type"`
	Fid         int         `json:"fid"`
	Timestamp   int64       `json:"timestamp"`
	Network     string      `json:"network"`
	CastAddBody CastAddBody `json:"castAddBody"`
}

type CastResponse struct {
	Data            MessageData `json:"data"`
	Hash            string      `json:"hash"`
	HashScheme      string      `json:"hashScheme"`
	Signature       string      `json:"signature"`
	SignatureScheme string      `json:"signatureScheme"`
	Signer          string      `json:"signer"`
}
