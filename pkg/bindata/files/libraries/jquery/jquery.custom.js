$(function() {
	$.widget('custom.iconselectmenu', $.ui.selectmenu, {
		_renderItem: function( ul, item ) {
			var li = $('<li>'),
			  wrapper = $('<div>', { text: item.label });

			if (item.disabled) {
			  li.addClass('ui-state-disabled');
			}

			$('<span>', {
			  style: item.element.attr('data-style'),
			  'class': 'ui-icon ' + item.element.attr('data-class')
			})
			  .appendTo(wrapper);

			return li.append(wrapper).appendTo(ul);
		}
	});
	
	if ($('#indexerReactionOptions').length) {
		$('#indexerReactionOptions').iconselectmenu({
			change: function(event, data) {
				insertIndexerReaction(data.item.value);
				$('#indexerReactionOptions').val(0);
				$('#indexerReactionOptions').iconselectmenu('refresh');
			}			
		}).iconselectmenu('menuWidget').addClass('ui-menu-icons customicons');
	}
});