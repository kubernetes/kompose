/* ---- Menu Section For Fold Personal Portfolio HTML Template --- */

jQuery(document).ready(function($){
	var $lateral_menu_trigger = $('.nav-trigger'),
		$content_wrapper = $('.main');

	//open-close lateral menu clicking on the menu icon
	$lateral_menu_trigger.on('click', function(event){
		event.preventDefault();

		$lateral_menu_trigger.toggleClass('is-active');
		$content_wrapper.toggleClass('is-active').one('webkitTransitionEnd otransitionend oTransitionEnd msTransitionEnd transitionend', function(){
			// firefox transitions break when parent overflow is changed, so we need to wait for the end of the trasition to give the body an overflow hidden
			$('body').toggleClass('overflow-hidden');
		});
		$('#fld-nav').toggleClass('is-active');

		//check if transitions are not supported - i.e. in IE9
		if($('html').hasClass('no-csstransitions')) {
			$('body').toggleClass('overflow-hidden');
		}
	});

	//close lateral menu clicking outside the menu itself
	$content_wrapper.on('mouseover', function(event){
		if( !$(event.target).is('.nav-trigger, .nav-trigger span') ) {
			$lateral_menu_trigger.removeClass('is-active');
			$content_wrapper.removeClass('is-active').one('webkitTransitionEnd otransitionend oTransitionEnd msTransitionEnd transitionend', function(){
				$('body').removeClass('overflow-hidden');
			});
			$('#fld-nav').removeClass('is-active');
			//check if transitions are not supported
			if($('html').hasClass('no-csstransitions')) {
				$('body').removeClass('overflow-hidden');
			}

		}
	});
});
