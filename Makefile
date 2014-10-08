MAN_DIR  := ./man
MAN_PAGE := $(MAN_DIR)/fzz.1

$(MAN_DIR)/%.1: $(MAN_DIR)/%.markdown
	@which md2man-roff >/dev/null || (echo "md2man missing: gem install md2man"; exit 1)
	md2man-roff $< > $@

man: $(MAN_PAGE)
